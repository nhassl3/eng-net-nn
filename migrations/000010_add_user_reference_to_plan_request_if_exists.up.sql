CREATE TABLE IF NOT EXISTS link_user_with_plan (
    user_id uuid not null,
    plan_id uuid unique not null
);

ALTER TABLE "link_user_with_plan" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE;
ALTER TABLE "link_user_with_plan" ADD FOREIGN KEY ("plan_id") REFERENCES "plans" ("id") ON DELETE CASCADE;