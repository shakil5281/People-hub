-- FINAL MERGE: After both geo and postcode seeds, merge all duplicates

-- 1. Merge duplicate divisions by name_bn (Bengali name is the canonical key)
DO $$
DECLARE
    rec RECORD;
    kept_id UUID;
    dup_id UUID;
BEGIN
    FOR rec IN
        SELECT name_bn,
               ARRAY_AGG(id ORDER BY LENGTH(name), id) as ids
        FROM divisions
        GROUP BY name_bn
        HAVING COUNT(*) > 1
    LOOP
        kept_id := rec.ids[1];
        FOR dup_id IN SELECT UNNEST(rec.ids[2:])
        LOOP
            UPDATE districts SET division_id = kept_id WHERE division_id = dup_id;
            DELETE FROM divisions WHERE id = dup_id;
        END LOOP;
    END LOOP;
END $$;

-- 2. Merge duplicate districts by name_bn
DO $$
DECLARE
    rec RECORD;
    kept_id UUID;
    dup_id UUID;
BEGIN
    FOR rec IN
        SELECT name_bn,
               ARRAY_AGG(id ORDER BY (
                   SELECT COUNT(*) FROM upazilas WHERE district_id = districts.id
               ) DESC, id) as ids
        FROM districts
        GROUP BY name_bn
        HAVING COUNT(*) > 1
    LOOP
        kept_id := rec.ids[1];
        FOR dup_id IN SELECT UNNEST(rec.ids[2:])
        LOOP
            UPDATE upazilas SET district_id = kept_id WHERE district_id = dup_id;
            DELETE FROM districts WHERE id = dup_id;
        END LOOP;
    END LOOP;
END $$;

-- 3. Remove orphaned upazilas
DELETE FROM upazilas WHERE district_id NOT IN (SELECT id FROM districts);

-- 4. Remove duplicate upazilas (same name + same district)
DELETE FROM upazilas WHERE id IN (
    SELECT id FROM (
        SELECT id, ROW_NUMBER() OVER (PARTITION BY name, district_id ORDER BY id) as rn
        FROM upazilas
    ) t WHERE rn > 1
);

-- 5. Remove orphaned unions
DELETE FROM unions WHERE upazila_id NOT IN (SELECT id FROM upazilas);

-- 6. Remove duplicate unions (same name + same upazila)
DELETE FROM unions WHERE id IN (
    SELECT id FROM (
        SELECT id, ROW_NUMBER() OVER (PARTITION BY name, upazila_id ORDER BY id) as rn
        FROM unions
    ) t WHERE rn > 1
);

-- 7. Fix post_offices with invalid district references
-- Match post_offices to districts by checking if the district still exists
DELETE FROM post_offices WHERE district_id NOT IN (SELECT id FROM districts);

-- Final counts
SELECT 'divisions' as table_name, COUNT(*) as count FROM divisions
UNION ALL SELECT 'districts', COUNT(*) FROM districts
UNION ALL SELECT 'upazilas', COUNT(*) FROM upazilas
UNION ALL SELECT 'unions', COUNT(*) FROM unions
UNION ALL SELECT 'post_offices', COUNT(*) FROM post_offices;
