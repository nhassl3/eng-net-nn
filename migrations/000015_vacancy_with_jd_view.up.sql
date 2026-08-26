CREATE VIEW vacancy_with_jd AS
SELECT v.id, v.jd, v.name, v.description, v.required_exp, v.pay_day, v.skills, v.created_at, v.updated_at,
       jd.name as jd_name, jd.tags as jd_tags, jd.description as jd_description
FROM vacancies v JOIN job_directions jd ON v.jd = jd.id;
