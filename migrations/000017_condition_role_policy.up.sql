CREATE OR REPLACE FUNCTION check_single_role() RETURNS trigger AS $$
       BEGIN
        PERFORM 1 FROM users WHERE id=NEW.user_id FOR UPDATE;

        IF TG_TABLE_NAME='admins' THEN
           IF EXISTS (SELECT 1 FROM partners WHERE user_id=NEW.user_id) THEN
            RAISE EXCEPTION 'user % is already a partner', NEW.user_id;
           END IF;
        ELSIF TG_TABLE_NAME='partners' THEN
              IF EXISTS (SELECT 1 FROM admins WHERE user_id=NEW.user_id) THEN
                RAISE EXCEPTION 'user % is already an admin', NEW.user_id;
              END IF;
        END IF;
        RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS admins_single_role ON admins;
DROP TRIGGER IF EXISTS partners_single_role ON partners;

CREATE TRIGGER admins_single_role
    BEFORE INSERT OR UPDATE ON admins
    FOR EACH ROW EXECUTE FUNCTION check_single_role();

CREATE TRIGGER partners_single_role
    BEFORE INSERT OR UPDATE ON partners
    FOR EACH ROW EXECUTE FUNCTION check_single_role();