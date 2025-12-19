# Chat Integration with Model Context - Implementation Plan

## Executive Summary

**Good News:** Your backend is complete and well-architected. The Claude Sonnet 4.5 integration is fully implemented with excellent context management. The frontend message display works perfectly. The **only missing piece** is wiring up the message input UI (textarea + send button).

**Current Context Assessment:** Already excellent and sufficient
- ✅ Last 10 messages for conversation history
- ✅ User preferences (year, make, model)
- ✅ Thread metadata (seller name, type)
- ✅ System prompt with negotiation guidelines

## Implementation Phases

### Phase 1: Wire Up Message Input (CRITICAL - Primary Work)

**Goal:** Enable users to send messages and receive AI-enhanced responses

**File:** [frontend/components/dashboard/ChatPane.tsx](frontend/components/dashboard/ChatPane.tsx)

**Current Issue:**
- Lines 258-276: Message input UI exists but has NO handlers
- No `onChange` on textarea
- No `onClick` on send button
- No state management for input value

**Changes Required:**

#### 1. Add State Variables (after line 17)
```typescript
const [messageInput, setMessageInput] = useState('');
const [sendingMessage, setSendingMessage] = useState(false);
const [sendError, setSendError] = useState<string | null>(null);
const messagesEndRef = useRef<HTMLDivElement>(null);
```

#### 2. Add Imports
```typescript
import { useRef } from 'react'; // Add to existing import
```

#### 3. Create Message Send Handler
```typescript
const handleSendMessage = async () => {
  if (!selectedThreadId || !messageInput.trim() || sendingMessage) {
    return;
  }

  const content = messageInput.trim();
  setSendingMessage(true);
  setSendError(null);

  try {
    const response = await messageAPI.createMessage(selectedThreadId, {
      content,
      sender: 'user'
    });

    // Add both user and agent messages to local state
    setMessages(prev => [
      ...prev,
      response.userMessage,
      ...(response.agentMessage ? [response.agentMessage] : [])
    ]);

    // Clear input
    setMessageInput('');
  } catch (error) {
    console.error('Failed to send message:', error);
    setSendError('Failed to send message. Please try again.');
  } finally {
    setSendingMessage(false);
  }
};
```

#### 4. Add Keyboard Handler
```typescript
const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
  if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
    e.preventDefault();
    handleSendMessage();
  }
};
```

#### 5. Add Auto-Scroll Effect
```typescript
useEffect(() => {
  messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
}, [messages]);
```

#### 6. Update Textarea (replace lines 262-266)
```typescript
<textarea
  value={messageInput}
  onChange={(e) => setMessageInput(e.target.value)}
  onKeyDown={handleKeyDown}
  placeholder="Type message... AI will assist (Ctrl+Enter to send)"
  rows={3}
  disabled={sendingMessage}
  className="w-full px-4 py-3 border border-gray-300 rounded-lg resize-none focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:opacity-50 disabled:cursor-not-allowed"
/>
```

#### 7. Update Send Button (replace lines 268-273)
```typescript
<button
  onClick={handleSendMessage}
  disabled={sendingMessage || !messageInput.trim()}
  className="bg-blue-600 hover:bg-blue-700 text-white px-6 py-3 rounded-lg font-medium transition-colors flex items-center gap-2 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
>
  {sendingMessage ? (
    <>
      <span>SENDING...</span>
      <svg className="animate-spin w-5 h-5" fill="none" viewBox="0 0 24 24">
        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
      </svg>
    </>
  ) : (
    <>
      <span>SEND</span>
      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
      </svg>
    </>
  )}
</button>
```

#### 8. Add Loading Indicator in Messages Area
After the messages.map() closing, before line 255:
```typescript
{sendingMessage && (
  <div className="flex justify-start">
    <div className="max-w-[70%]">
      <div className="text-xs text-gray-500 mb-1">AI Agent</div>
      <div className="rounded-lg px-4 py-3 bg-purple-100 text-gray-800 border border-purple-200">
        <div className="flex items-center gap-2">
          <svg className="animate-spin w-4 h-4 text-purple-600" fill="none" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          <span className="text-sm">Crafting negotiation message...</span>
        </div>
      </div>
    </div>
  </div>
)}
```

#### 9. Add Scroll Target
After messages area, before closing div:
```typescript
<div ref={messagesEndRef} />
```

#### 10. Add Error Display (Optional but Recommended)
After send button, before closing div:
```typescript
{sendError && (
  <div className="absolute bottom-full mb-2 right-0 bg-red-100 border border-red-400 text-red-700 px-4 py-2 rounded-lg text-sm">
    {sendError}
  </div>
)}
```

---

### Phase 2: Testing (CRITICAL - Validation)

**Manual Test Checklist:**
- [ ] Login and select a thread
- [ ] Type message in textarea and click send
- [ ] Verify user message appears
- [ ] Verify loading indicator shows
- [ ] Verify AI agent response appears (enhanced version)
- [ ] Verify both messages persist after page refresh
- [ ] Test Ctrl+Enter keyboard shortcut
- [ ] Test empty message prevention (button should be disabled)
- [ ] Test error handling (stop backend, try to send)
- [ ] Switch between threads and verify context isolation

**API Test:**
```bash
# Test message creation endpoint
curl -X POST http://localhost:8080/api/v1/threads/{thread-id}/messages \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{"content": "What is your best price?", "sender": "user"}'
```

Expected response includes both userMessage and agentMessage.

---

### Phase 3: Optional Enhancements (POST-MVP)

#### Enhancement 1: Add Tracked Offers to Context (Medium Priority)

**Purpose:** Enable AI to reference offers from other threads for competitive negotiation

**Files to Modify:**

1. **[backend/internal/services/message.go](backend/internal/services/message.go:59-106)** - CreateUserMessage function
   - After line 87, fetch tracked offers:
   ```go
   var trackedOffers []models.TrackedOffer
   s.db.Joins("JOIN threads ON tracked_offers.thread_id = threads.id").
       Where("threads.user_id = ?", userID).
       Order("tracked_offers.tracked_at DESC").
       Limit(5).
       Find(&trackedOffers)
   ```
   - Pass to Claude service

2. **[backend/internal/services/claude.go](backend/internal/services/claude.go:34-54)** - GenerateNegotiationResponse
   - Add trackedOffers parameter
   - Append to system prompt:
   ```go
   if len(trackedOffers) > 0 {
       systemPrompt += "\n\nTracked Offers Across All Negotiations:\n"
       for _, offer := range trackedOffers {
           systemPrompt += fmt.Sprintf("- %s\n", offer.OfferText)
       }
       systemPrompt += "\nYou can reference these offers when negotiating for competitive advantage."
   }
   ```

#### Enhancement 2: Display User Preferences in Sidebar (Low Priority)

**File:** [frontend/components/dashboard/ThreadPane.tsx](frontend/components/dashboard/ThreadPane.tsx)

Add after line 110:
```typescript
<div className="border-b border-slate-700 px-4 py-3">
  <div className="text-xs text-slate-400 uppercase mb-1">Target Vehicle</div>
  <div className="text-sm text-white font-medium">
    {user?.preferences?.year} {user?.preferences?.make} {user?.preferences?.model}
  </div>
</div>
```

#### Enhancement 3: Increase Message History Limit (Optional)

**Current:** 10 messages for Claude context
**Recommendation:** Keep at 10, or increase to maximum 20

**File:** [backend/internal/services/message.go](backend/internal/services/message.go:82)
```go
// Change Limit(10) to Limit(20) if needed
s.db.Where("thread_id = ?", threadID).Order("timestamp DESC").Limit(20).Find(&recentMessages)
```

**Note:** Only increase if you find 10 messages insufficient. More context = higher API costs and slower responses.

---

## Critical Files Reference

### Files to Modify (Phase 1 - Required):
1. [frontend/components/dashboard/ChatPane.tsx](frontend/components/dashboard/ChatPane.tsx) - PRIMARY WORK HERE

### Files to Modify (Phase 3 - Optional):
1. [backend/internal/services/message.go](backend/internal/services/message.go)
2. [backend/internal/services/claude.go](backend/internal/services/claude.go)
3. [frontend/components/dashboard/ThreadPane.tsx](frontend/components/dashboard/ThreadPane.tsx)

### Files That Work Perfectly (NO CHANGES NEEDED):
1. [backend/internal/api/handlers/message.go](backend/internal/api/handlers/message.go) - API endpoints complete
2. [frontend/lib/api.ts](frontend/lib/api.ts) - API client ready
3. [backend/internal/db/models/message.go](backend/internal/db/models/message.go) - Database models solid

---

## Context Management Analysis

### Current Context (Excellent and Sufficient)

**What Claude receives per message:**
1. **System Prompt:**
   - User's car requirements (year, make, model)
   - Seller name
   - Negotiation guidelines

2. **Message History:**
   - Last 10 messages from thread
   - Formatted with sender labels (user/agent/seller)

3. **Current Message:**
   - User's draft message to enhance

**Assessment:** This is well-designed and provides sufficient context for effective negotiation assistance. The 10-message limit balances context quality with API efficiency.

### Why Current Context Is Good Enough

- ✅ **10 messages = ~5 conversation turns** - Sufficient for coherent conversation
- ✅ **User preferences included** - AI knows what car they want
- ✅ **Seller context present** - Knows who they're negotiating with
- ✅ **Clear AI role** - Enhance messages for negotiation
- ✅ **Cost-efficient** - Doesn't waste tokens on excessive history
- ✅ **Fast responses** - Less context = faster Claude API calls

---

## Implementation Strategy

### Recommended Path

**Step 1 (Today - 2 hours):**
- Implement Phase 1: Wire up message input in ChatPane
- Test basic send/receive flow
- Validate end-to-end with real messages

**Step 2 (Testing - 30 minutes):**
- Run through manual test checklist
- Verify AI responses are contextually appropriate
- Test error scenarios

**Step 3 (Optional - Later):**
- Add tracked offers context if competitive negotiation is needed
- Display user preferences in sidebar for visibility
- Consider message history adjustments based on usage

### What NOT to Do

- ❌ Don't redesign backend Claude integration (it's excellent)
- ❌ Don't add WebSockets/real-time updates (not needed for MVP)
- ❌ Don't implement message editing/deletion (post-MVP)
- ❌ Don't add file attachments (out of scope)
- ❌ Don't over-engineer context management (current is optimal)

---

## Expected Results

**After Phase 1:**
- Users can type and send messages
- AI agent receives user input and enhances it for negotiation
- Both user and agent messages display in chat
- Conversation flows naturally
- Context is maintained across the conversation
- Loading states provide clear feedback

**Example Flow:**
1. User types: "price?"
2. User clicks Send
3. User message appears: "price?"
4. Loading indicator shows: "Crafting negotiation message..."
5. Agent message appears: "We are serious buyers for this 2024 Mazda CX-90. Could you please provide your best out-the-door pricing, including all fees and taxes?"
6. Both messages saved to database and persist

---

## Success Criteria

✅ Users can send messages via textarea and send button
✅ AI agent responds with enhanced negotiation message
✅ Loading states show during AI generation
✅ Errors are handled gracefully
✅ Messages persist across page refreshes
✅ Context is maintained within threads
✅ Keyboard shortcuts work (Ctrl+Enter)
✅ Auto-scroll to new messages
✅ Thread isolation works correctly
