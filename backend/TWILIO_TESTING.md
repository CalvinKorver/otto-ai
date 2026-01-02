# Twilio SMS Integration Testing Guide

## Prerequisites

1. **Twilio Account Setup**
   - Create a Twilio account at https://www.twilio.com
   - Get your Account SID and Auth Token from the Twilio Console
   - Create a Messaging Service in Twilio Console
   - Get the Messaging Service SID
   - Set up A2P 10DLC Brand and Campaign (required for production, optional for testing)

2. **Environment Variables**
   Add these to your `.env` file:
   ```bash
   TWILIO_ACCOUNT_SID=ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
   TWILIO_AUTH_TOKEN=your_auth_token_here
   TWILIO_MESSAGING_SERVICE_SID=MGxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
   TWILIO_WEBHOOK_URL=https://your-domain.com/api/v1/webhooks/sms/inbound
   ```

## Testing Steps

### 1. Test Phone Number Allocation

**Goal**: Verify that phone numbers are automatically allocated when users complete onboarding.

**Steps**:
1. Start your backend server
2. Register a new user or complete onboarding for an existing user
3. Check the database to verify:
   ```sql
   SELECT id, email, phone_number, twilio_sid FROM users WHERE phone_number IS NOT NULL;
   ```
4. Verify in Twilio Console:
   - Go to Phone Numbers → Manage → Active numbers
   - You should see the newly purchased number
   - Go to Messaging → Services → Your Service → Phone Numbers
   - The number should be listed there (assigned to Messaging Service)

**Expected Result**: 
- User record has `phone_number` and `twilio_sid` populated
- Number appears in Twilio Console
- Number is assigned to your Messaging Service

### 2. Test Incoming SMS Webhook (Local Development)

**Goal**: Verify that incoming SMS messages are received and stored correctly.

**Setup for Local Testing**:
1. Use ngrok or similar tool to expose your local server:
   ```bash
   ngrok http 8080
   ```
2. Update Twilio webhook URL:
   - Go to Twilio Console → Messaging → Services → Your Service
   - Set "Inbound Message" webhook to: `https://your-ngrok-url.ngrok.io/api/v1/webhooks/sms/inbound`
   - Save

**Steps**:
1. Send a test SMS to your allocated phone number from your personal phone
2. Check backend logs for webhook receipt
3. Verify in database:
   ```sql
   SELECT id, sender, content, sender_phone, message_type_id, thread_id 
   FROM messages 
   WHERE sender_phone IS NOT NULL 
   ORDER BY timestamp DESC 
   LIMIT 5;
   ```
4. Check the inbox in your frontend - the SMS should appear as an unassigned message

**Expected Result**:
- Webhook is received (check logs)
- Message is created in database with `sender_phone` populated
- Message has `thread_id = NULL` (unassigned, in inbox)
- Message appears in frontend inbox

### 3. Test SMS Assignment to Thread

**Goal**: Verify that assigning an SMS to a thread sets the thread's phone number.

**Steps**:
1. Create a thread (or use existing)
2. Assign an inbox SMS message to that thread via the frontend
3. Verify in database:
   ```sql
   SELECT id, seller_name, phone FROM threads WHERE phone IS NOT NULL;
   ```
4. The thread should now have the dealer's phone number

**Expected Result**:
- Thread's `phone` field is set to the SMS sender's phone number
- Message's `thread_id` is updated

### 4. Test Outgoing SMS

**Goal**: Verify that sending SMS replies works correctly.

**Steps**:
1. In the frontend, open a thread that has a phone number assigned
2. View an agent message
3. Click "Send SMS" button
4. Confirm the dialog
5. Check backend logs for Twilio API call
6. Verify in database:
   ```sql
   SELECT id, sender, content, sent_via_sms, sender_phone 
   FROM messages 
   WHERE sent_via_sms = true 
   ORDER BY timestamp DESC 
   LIMIT 5;
   ```
7. Check your phone - you should receive the SMS (if using a real number)

**Expected Result**:
- SMS is sent via Twilio API (check logs)
- Message is saved with `sent_via_sms = true`
- Message appears in thread with "✔️ Sent as SMS" indicator
- SMS is received on the dealer's phone

### 5. Test Webhook Signature Validation

**Goal**: Verify that invalid webhook requests are rejected.

**Steps**:
1. Send a POST request to `/api/v1/webhooks/sms/inbound` without the `X-Twilio-Signature` header:
   ```bash
   curl -X POST http://localhost:8080/api/v1/webhooks/sms/inbound \
     -H "Content-Type: application/x-www-form-urlencoded" \
     -d "To=+1234567890&From=+0987654321&Body=Test"
   ```
2. Should receive 401 Unauthorized (if auth token is set)

**Expected Result**: Invalid requests are rejected

### 6. Test Error Handling

**Test Scenarios**:
1. **No available numbers**: Try allocating when Twilio has no numbers available
2. **Invalid area code**: Test with invalid area code
3. **Twilio API failure**: Temporarily use wrong credentials
4. **User already has number**: Try allocating again (should be idempotent)

## Testing with Twilio Test Credentials

For development/testing, you can use Twilio's test credentials:

```bash
# Test credentials (these are public and safe to use)
TWILIO_ACCOUNT_SID=ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
TWILIO_AUTH_TOKEN=test_token_here
```

**Note**: Test credentials have limitations:
- Cannot purchase real numbers
- Cannot send to real phone numbers
- Use for API structure testing only

## Manual Testing Checklist

- [ ] Phone number allocated during onboarding
- [ ] Number appears in Twilio Console
- [ ] Number is assigned to Messaging Service
- [ ] Incoming SMS creates inbox message
- [ ] Inbox SMS can be assigned to thread
- [ ] Thread phone number is set when SMS assigned
- [ ] "Send SMS" button appears for agent messages
- [ ] SMS confirmation dialog works
- [ ] SMS is sent successfully
- [ ] "Sent as SMS" indicator appears
- [ ] Phone number displays on dashboard
- [ ] Webhook signature validation works

## Debugging Tips

1. **Check Backend Logs**: Look for Twilio API errors or webhook processing logs
2. **Twilio Console**: Check the Twilio Console for API logs and message status
3. **Database Queries**: Use SQL queries to verify data is being stored correctly
4. **Network**: Use ngrok logs to see incoming webhook requests
5. **Frontend Console**: Check browser console for API errors

## Common Issues

1. **Number not assigned to Messaging Service**: Check that `TWILIO_MESSAGING_SERVICE_SID` is set correctly
2. **Webhook not receiving**: Verify webhook URL is accessible and correctly configured in Twilio
3. **SMS not sending**: Check that Messaging Service has A2P Campaign linked
4. **Signature validation failing**: Ensure `TWILIO_AUTH_TOKEN` matches the one in Twilio Console

## Production Readiness Checklist

Before going to production:
- [ ] A2P 10DLC Brand registered
- [ ] A2P 10DLC Campaign registered and approved
- [ ] Messaging Service linked to Campaign
- [ ] Webhook URL is publicly accessible (HTTPS)
- [ ] Webhook signature validation enabled
- [ ] Error handling and retry logic tested
- [ ] Monitoring/alerting set up for Twilio API failures

