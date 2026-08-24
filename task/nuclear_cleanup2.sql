-- NUCLEAR CLEANUP ROUND 2: Delete everything and start fresh
DELETE FROM unions;
DELETE FROM upazilas;
DELETE FROM districts;
DELETE FROM divisions;
DELETE FROM post_offices;

SELECT 'All tables cleared' as status;
