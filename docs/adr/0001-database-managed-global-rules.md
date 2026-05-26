# Database-Managed Global Rules

Classification rules that apply to every Mailbox Owner are stored as Global Rules in `user_rules` and seeded by SQL migrations. The classifier does not maintain a separate built-in rule list in application code because the database should be the source of truth for rules. Future changes can be made through migrations, direct database edits in local development, or a rules management UI if one is added later.

Rules do not have a separate stable `rule_key`; duplicate protection is based on the meaningful rule fields instead. This keeps the rule model small while still preventing accidental duplicate Global Rules or User-Specific Rules. The deployment targets PostgreSQL 16, so the duplicate-protection index can use `NULLS NOT DISTINCT` to treat Global Rules with `user_id = NULL` as comparable for uniqueness.

User-Specific Rules take precedence over Global Rules because they represent the Mailbox Owner's explicit override for their inbox. Rule priority still orders rules within the same scope, but it does not let a Global Rule beat a matching User-Specific Rule.
