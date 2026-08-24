-- Remove remaining duplicate districts (keep the one with more upazila children)
DO $$
DECLARE
    rec RECORD;
    kept_id UUID;
    dup_id UUID;
    dup_ids UUID[];
BEGIN
    FOR rec IN
        SELECT name, name_bn,
               ARRAY_AGG(id ORDER BY (
                   SELECT COUNT(*) FROM upazilas WHERE district_id = districts.id
               ) DESC) as ids
        FROM districts
        GROUP BY name, name_bn
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

-- Remove orphaned upazilas
DELETE FROM upazilas WHERE district_id NOT IN (SELECT id FROM districts);

-- Remove duplicate upazilas
DELETE FROM upazilas WHERE id IN (
    SELECT id FROM (
        SELECT id, ROW_NUMBER() OVER (PARTITION BY name, district_id ORDER BY id) as rn
        FROM upazilas
    ) t WHERE rn > 1
);

-- Remove orphaned unions
DELETE FROM unions WHERE upazila_id NOT IN (SELECT id FROM upazilas);

-- Remove duplicate unions
DELETE FROM unions WHERE id IN (
    SELECT id FROM (
        SELECT id, ROW_NUMBER() OVER (PARTITION BY name, upazila_id ORDER BY id) as rn
        FROM unions
    ) t WHERE rn > 1
);

SELECT 'divisions' as tbl, COUNT(*) FROM divisions
UNION ALL SELECT 'districts', COUNT(*) FROM districts
UNION ALL SELECT 'upazilas', COUNT(*) FROM upazilas
UNION ALL SELECT 'unions', COUNT(*) FROM unions;
