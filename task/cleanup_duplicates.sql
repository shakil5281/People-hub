-- Step 1: Remove duplicate districts (keep the one with most upazila children)
DO $$
DECLARE
    rec RECORD;
    kept_id UUID;
    dup_id UUID;
BEGIN
    FOR rec IN
        SELECT name,
               ARRAY_AGG(id ORDER BY (
                   SELECT COUNT(*) FROM upazilas WHERE district_id = districts.id
               ) DESC) as ids
        FROM districts
        GROUP BY name
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

-- Step 2: Merge naming variants
DO $$
BEGIN
    -- Cumilla -> Comilla
    IF EXISTS (SELECT 1 FROM districts WHERE name='Cumilla') AND EXISTS (SELECT 1 FROM districts WHERE name='Comilla') THEN
        UPDATE upazilas SET district_id = (SELECT id FROM districts WHERE name='Comilla') WHERE district_id IN (SELECT id FROM districts WHERE name='Cumilla');
        DELETE FROM districts WHERE name='Cumilla';
    END IF;
    -- Chapainawabganj -> Chapai Nawabganj
    IF EXISTS (SELECT 1 FROM districts WHERE name='Chapainawabganj') AND EXISTS (SELECT 1 FROM districts WHERE name='Chapai Nawabganj') THEN
        UPDATE upazilas SET district_id = (SELECT id FROM districts WHERE name='Chapai Nawabganj') WHERE district_id IN (SELECT id FROM districts WHERE name='Chapainawabganj');
        DELETE FROM districts WHERE name='Chapainawabganj';
    END IF;
    -- Khagrachhari -> Khagrachari
    IF EXISTS (SELECT 1 FROM districts WHERE name='Khagrachhari') AND EXISTS (SELECT 1 FROM districts WHERE name='Khagrachari') THEN
        UPDATE upazilas SET district_id = (SELECT id FROM districts WHERE name='Khagrachari') WHERE district_id IN (SELECT id FROM districts WHERE name='Khagrachhari');
        DELETE FROM districts WHERE name='Khagrachhari';
    END IF;
    -- Netrakona -> Netrokona
    IF EXISTS (SELECT 1 FROM districts WHERE name='Netrakona') AND EXISTS (SELECT 1 FROM districts WHERE name='Netrokona') THEN
        UPDATE upazilas SET district_id = (SELECT id FROM districts WHERE name='Netrokona') WHERE district_id IN (SELECT id FROM districts WHERE name='Netrakona');
        DELETE FROM districts WHERE name='Netrakona';
    END IF;
    -- Jhalokati -> Jhalakathi
    IF EXISTS (SELECT 1 FROM districts WHERE name='Jhalokati') AND EXISTS (SELECT 1 FROM districts WHERE name='Jhalakathi') THEN
        UPDATE upazilas SET district_id = (SELECT id FROM districts WHERE name='Jhalakathi') WHERE district_id IN (SELECT id FROM districts WHERE name='Jhalokati');
        DELETE FROM districts WHERE name='Jhalokati';
    END IF;
    -- Barishal -> Barisal
    IF EXISTS (SELECT 1 FROM districts WHERE name='Barishal') AND EXISTS (SELECT 1 FROM districts WHERE name='Barisal') THEN
        UPDATE upazilas SET district_id = (SELECT id FROM districts WHERE name='Barisal') WHERE district_id IN (SELECT id FROM districts WHERE name='Barishal');
        DELETE FROM districts WHERE name='Barishal';
    END IF;
    -- Coxsbazar -> Cox's Bazar (rename the one without apostrophe)
    IF EXISTS (SELECT 1 FROM districts WHERE name='Coxsbazar') AND EXISTS (SELECT 1 FROM districts WHERE name='Cox''s Bazar') THEN
        UPDATE upazilas SET district_id = (SELECT id FROM districts WHERE name='Cox''s Bazar') WHERE district_id IN (SELECT id FROM districts WHERE name='Coxsbazar');
        DELETE FROM districts WHERE name='Coxsbazar';
    END IF;
END $$;

-- Step 3: Remove orphaned upazilas (upazilas pointing to non-existent districts)
DELETE FROM upazilas WHERE district_id NOT IN (SELECT id FROM districts);

-- Step 4: Remove duplicate upazilas within same district (keep the one with more unions)
DO $$
DECLARE
    rec RECORD;
    kept_id UUID;
    dup_id UUID;
BEGIN
    FOR rec IN
        SELECT name, district_id,
               ARRAY_AGG(id ORDER BY (
                   SELECT COUNT(*) FROM unions WHERE upazila_id = upazilas.id
               ) DESC) as ids
        FROM upazilas
        GROUP BY name, district_id
        HAVING COUNT(*) > 1
    LOOP
        kept_id := rec.ids[1];
        FOR dup_id IN SELECT UNNEST(rec.ids[2:])
        LOOP
            UPDATE unions SET upazila_id = kept_id WHERE upazila_id = dup_id;
            DELETE FROM upazilas WHERE id = dup_id;
        END LOOP;
    END LOOP;
END $$;

-- Step 5: Remove orphaned unions
DELETE FROM unions WHERE upazila_id NOT IN (SELECT id FROM upazilas);

-- Step 6: Update post_offices district_id to match cleaned districts
UPDATE post_offices SET district_id = (
    SELECT d.id FROM districts d WHERE d.name = post_offices.district_id::text
) WHERE district_id IS NOT NULL AND district_id::text NOT IN (SELECT id::text FROM districts);

-- Final counts
SELECT 'divisions' as tbl, COUNT(*) FROM divisions
UNION ALL SELECT 'districts', COUNT(*) FROM districts
UNION ALL SELECT 'upazilas', COUNT(*) FROM upazilas
UNION ALL SELECT 'unions', COUNT(*) FROM unions
UNION ALL SELECT 'post_offices', COUNT(*) FROM post_offices;
