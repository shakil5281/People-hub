-- More aggressive dedup: for districts with same name AND same name_bn, keep one
DO $$
DECLARE
    rec RECORD;
    kept_id UUID;
    dup_id UUID;
BEGIN
    FOR rec IN
        SELECT name, name_bn,
               ARRAY_AGG(id ORDER BY id) as ids
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

-- Re-clean upazilas
DELETE FROM upazilas WHERE district_id NOT IN (SELECT id FROM districts);
DELETE FROM upazilas WHERE id IN (
    SELECT id FROM (
        SELECT id, ROW_NUMBER() OVER (PARTITION BY name, district_id ORDER BY id) as rn
        FROM upazilas
    ) t WHERE rn > 1
);

-- Re-clean unions
DELETE FROM unions WHERE upazila_id NOT IN (SELECT id FROM upazilas);
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
