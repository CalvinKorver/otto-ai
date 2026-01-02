-- Clean Database Script
-- This script removes conversation data while preserving:
-- - User accounts (users, user_preferences, gmail_tokens)
-- - Reference data (makes, message_types, models, vehicle_trims)
--
-- Tables that will be cleared:
-- - threads
-- - messages
-- - tracked_offers
-- - dealers
--
-- Tables that will be preserved:
-- - users
-- - user_preferences
-- - gmail_tokens
-- - makes
-- - message_types
-- - models
-- - vehicle_trims
--
-- Usage:
--   psql -U your_user -d your_database -f clean-db.sql
--   Or from psql: \i clean-db.sql

BEGIN;

-- Show counts before deletion
DO $$
DECLARE
    user_count INTEGER;
    thread_count INTEGER;
    message_count INTEGER;
    offer_count INTEGER;
    dealer_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO user_count FROM users;
    SELECT COUNT(*) INTO thread_count FROM threads;
    SELECT COUNT(*) INTO message_count FROM messages;
    SELECT COUNT(*) INTO offer_count FROM tracked_offers;
    SELECT COUNT(*) INTO dealer_count FROM dealers;
    
    RAISE NOTICE 'Before cleanup:';
    RAISE NOTICE '  Users: % (preserved)', user_count;
    RAISE NOTICE '  Threads: % (will be deleted)', thread_count;
    RAISE NOTICE '  Messages: % (will be deleted)', message_count;
    RAISE NOTICE '  Tracked Offers: % (will be deleted)', offer_count;
    RAISE NOTICE '  Dealers: % (will be deleted)', dealer_count;
END $$;

-- Delete in order respecting foreign key constraints (child tables first)

DO $$
BEGIN
    -- 1. Delete tracked offers (references threads and messages)
    DELETE FROM tracked_offers;
    RAISE NOTICE 'Deleted tracked_offers';

    -- 2. Delete messages (references threads, users, message_types)
    DELETE FROM messages;
    RAISE NOTICE 'Deleted messages';

    -- 3. Delete threads (references users)
    DELETE FROM threads;
    RAISE NOTICE 'Deleted threads';

    -- 4. Delete dealers (references user_preferences)
    DELETE FROM dealers;
    RAISE NOTICE 'Deleted dealers';

    -- Note: users, user_preferences, and gmail_tokens are preserved
END $$;

-- Verify what remains
DO $$
DECLARE
    user_count INTEGER;
    user_pref_count INTEGER;
    make_count INTEGER;
    message_type_count INTEGER;
    model_count INTEGER;
    trim_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO user_count FROM users;
    SELECT COUNT(*) INTO user_pref_count FROM user_preferences;
    SELECT COUNT(*) INTO make_count FROM makes;
    SELECT COUNT(*) INTO message_type_count FROM message_types;
    SELECT COUNT(*) INTO model_count FROM models;
    SELECT COUNT(*) INTO trim_count FROM vehicle_trims;
    
    RAISE NOTICE '';
    RAISE NOTICE 'After cleanup - Preserved data:';
    RAISE NOTICE '  Users: %', user_count;
    RAISE NOTICE '  User Preferences: %', user_pref_count;
    RAISE NOTICE '  Makes: %', make_count;
    RAISE NOTICE '  Message Types: %', message_type_count;
    RAISE NOTICE '  Models: %', model_count;
    RAISE NOTICE '  Vehicle Trims: %', trim_count;
    RAISE NOTICE '';
    RAISE NOTICE 'Database cleaned successfully! Conversation data removed, user accounts preserved.';
END $$;

COMMIT;

