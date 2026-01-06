-- Drop application tables (CASCADE removes dependent constraints/indexes)
DROP TABLE IF EXISTS tasks CASCADE;
DROP TABLE IF EXISTS user_organization_assignments CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS organizations CASCADE;
DROP TABLE IF EXISTS sessions CASCADE;