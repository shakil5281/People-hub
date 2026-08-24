-- FINAL COMPREHENSIVE CLEANUP
-- This script removes ALL duplicates while preserving data integrity

-- 1. Remove duplicate divisions (keep 'Barisal' not 'Barishal', etc.)
DELETE FROM divisions WHERE id IN (
    SELECT id FROM (
        SELECT id, name,
               ROW_NUMBER() OVER (PARTITION BY name_bn ORDER BY LENGTH(name), id) as rn
        FROM divisions
    ) t WHERE rn > 1
);

-- Also remove 'Barishal' if 'Barisal' exists
DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM divisions WHERE name='Barisal') AND EXISTS (SELECT 1 FROM divisions WHERE name='Barishal') THEN
        UPDATE districts SET division_id = (SELECT id FROM divisions WHERE name='Barisal') WHERE division_id = (SELECT id FROM divisions WHERE name='Barishal');
        DELETE FROM divisions WHERE name='Barishal';
    END IF;
END $$;

-- 2. Remove duplicate districts
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
               ) DESC) as ids
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

-- Also merge naming variants by name_bn
DO $$ BEGIN
    -- Comilla / Cumilla
    IF EXISTS (SELECT 1 FROM districts WHERE name='Comilla') AND EXISTS (SELECT 1 FROM districts WHERE name='Cumilla') THEN
        UPDATE upazilas SET district_id = (SELECT id FROM districts WHERE name='Comilla') WHERE district_id IN (SELECT id FROM districts WHERE name='Cumilla');
        DELETE FROM districts WHERE name='Cumilla';
    END IF;
    -- Coxsbazar / Cox's Bazar
    IF EXISTS (SELECT 1 FROM districts WHERE name='Cox''s Bazar') AND EXISTS (SELECT 1 FROM districts WHERE name='Coxsbazar') THEN
        UPDATE upazilas SET district_id = (SELECT id FROM districts WHERE name='Cox''s Bazar') WHERE district_id IN (SELECT id FROM districts WHERE name='Coxsbazar');
        DELETE FROM districts WHERE name='Coxsbazar';
    END IF;
    -- Chapainawabganj / Chapai Nawabganj
    IF EXISTS (SELECT 1 FROM districts WHERE name='Chapai Nawabganj') AND EXISTS (SELECT 1 FROM districts WHERE name='Chapainawabganj') THEN
        UPDATE upazilas SET district_id = (SELECT id FROM districts WHERE name='Chapai Nawabganj') WHERE district_id IN (SELECT id FROM districts WHERE name='Chapainawabganj');
        DELETE FROM districts WHERE name='Chapainawabganj';
    END IF;
    -- Khagrachari / Khagrachhari
    IF EXISTS (SELECT 1 FROM districts WHERE name='Khagrachari') AND EXISTS (SELECT 1 FROM districts WHERE name='Khagrachhari') THEN
        UPDATE upazilas SET district_id = (SELECT id FROM districts WHERE name='Khagrachari') WHERE district_id IN (SELECT id FROM districts WHERE name='Khagrachhari');
        DELETE FROM districts WHERE name='Khagrachhari';
    END IF;
    -- Netrokona / Netrakona
    IF EXISTS (SELECT 1 FROM districts WHERE name='Netrokona') AND EXISTS (SELECT 1 FROM districts WHERE name='Netrakona') THEN
        UPDATE upazilas SET district_id = (SELECT id FROM districts WHERE name='Netrokona') WHERE district_id IN (SELECT id FROM districts WHERE name='Netrakona');
        DELETE FROM districts WHERE name='Netrakona';
    END IF;
    -- Jhalakathi / Jhalokati
    IF EXISTS (SELECT 1 FROM districts WHERE name='Jhalakathi') AND EXISTS (SELECT 1 FROM districts WHERE name='Jhalokati') THEN
        UPDATE upazilas SET district_id = (SELECT id FROM districts WHERE name='Jhalakathi') WHERE district_id IN (SELECT id FROM districts WHERE name='Jhalokati');
        DELETE FROM districts WHERE name='Jhalokati';
    END IF;
    -- Barisal / Barishal
    IF EXISTS (SELECT 1 FROM districts WHERE name='Barisal') AND EXISTS (SELECT 1 FROM districts WHERE name='Barishal') THEN
        UPDATE upazilas SET district_id = (SELECT id FROM districts WHERE name='Barisal') WHERE district_id IN (SELECT id FROM districts WHERE name='Barishal');
        DELETE FROM districts WHERE name='Barishal';
    END IF;
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

-- 7. Delete ALL post_offices and re-seed (district IDs have changed)
DELETE FROM post_offices;

-- Final counts
SELECT 'divisions' as table_name, COUNT(*) as count FROM divisions
UNION ALL SELECT 'districts', COUNT(*) FROM districts
UNION ALL SELECT 'upazilas', COUNT(*) FROM upazilas
UNION ALL SELECT 'unions', COUNT(*) FROM unions
UNION ALL SELECT 'post_offices', COUNT(*) FROM post_offices;
