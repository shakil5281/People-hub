-- NUCLEAR CLEANUP: Delete everything and re-seed from scratch
-- This ensures clean, consistent data

-- Delete in correct order (children first)
DELETE FROM unions;
DELETE FROM upazilas;
DELETE FROM districts;
DELETE FROM divisions;
DELETE FROM post_offices;

-- Verify empty
SELECT 'divisions' as tbl, COUNT(*) FROM divisions
UNION ALL SELECT 'districts', COUNT(*) FROM districts
UNION ALL SELECT 'upazilas', COUNT(*) FROM upazilas
UNION ALL SELECT 'unions', COUNT(*) FROM unions
UNION ALL SELECT 'post_offices', COUNT(*) FROM post_offices;
