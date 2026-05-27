DELETE FROM user_rules
WHERE user_id IS NULL
  AND (rule_type, operator, rule_value, target_label, priority) IN (
    ('sender_domain', 'contains', 'google.com', 'Security', 220),
    ('subject', 'contains', 'sign-in', 'Security', 210),
    ('any', 'contains', 'verification code', 'Security', 210),
    ('any', 'contains', 'otp', 'Security', 210),
    ('any', 'contains', '2fa', 'Security', 210),
    ('any', 'contains', 'authenticator', 'Security', 210),
    ('any', 'contains', 'lock your account', 'Security', 210),
    ('any', 'contains', 'если это не были вы', 'Security', 210),
    ('any', 'contains', 'вошла в систему', 'Security', 205),
    ('any', 'contains', 'подтвердить адрес эл. почты', 'Security', 205),
    ('sender_domain', 'contains', 'bank.com', 'Transactions', 180),
    ('any', 'contains', 'transaction', 'Transactions', 170),
    ('any', 'contains', 'payment', 'Transactions', 170),
    ('any', 'contains', 'invoice', 'Transactions', 170),
    ('any', 'contains', 'receipt', 'Transactions', 170),
    ('any', 'contains', 'card', 'Transactions', 160),
    ('any', 'contains', 'qr pay', 'Transactions', 160),
    ('any', 'contains', 'thank you for applying', 'Job', 180),
    ('any', 'contains', 'your application has been received', 'Job', 180),
    ('any', 'contains', 'we have received your application', 'Job', 180),
    ('any', 'contains', 'thank you for taking the time to apply', 'Job', 180),
    ('any', 'contains', 'talent acquisition team', 'Job', 175),
    ('any', 'contains', 'will not be proceeding at this time', 'Job', 175),
    ('any', 'contains', 'hiring team', 'Job', 170),
    ('any', 'contains', 'backend engineer', 'Job', 170),
    ('any', 'contains', 'applicant', 'Job', 165),
    ('any', 'contains', 'position', 'Job', 155),
    ('any', 'contains', 'role', 'Job', 150),
    ('any', 'contains', 'interview', 'Job', 150),
    ('any', 'contains', 'recruiter', 'Job', 150),
    ('any', 'contains', 'job', 'Job', 140),
    ('any', 'contains', 'opportunity', 'Job', 140),
    ('any', 'contains', 'vacancy', 'Job', 140),
    ('any', 'contains', 'resume', 'Job', 130),
    ('any', 'contains', 'subscription will end', 'Promo', 145),
    ('any', 'contains', 'renew', 'Promo', 140),
    ('any', 'contains', 'upgrade', 'Promo', 140),
    ('any', 'contains', 'keep your', 'Promo', 130),
    ('any', 'contains', 'sale', 'Promo', 120),
    ('any', 'contains', 'discount', 'Promo', 120),
    ('any', 'contains', 'promo', 'Promo', 120),
    ('any', 'contains', 'coupon', 'Promo', 120),
    ('any', 'contains', 'offer', 'Promo', 110),
    ('any', 'contains', 'deal', 'Promo', 110),
    ('any', 'contains', '% off', 'Promo', 110),
    ('any', 'contains', 'notification', 'Social', 100),
    ('any', 'contains', 'friend', 'Social', 100),
    ('any', 'contains', 'comment', 'Social', 100),
    ('any', 'contains', 'mention', 'Social', 100),
    ('any', 'contains', 'invitation', 'Social', 100)
  );

DELETE FROM user_rules
WHERE user_id = '00000000-0000-0000-0000-000000000001'
  AND rule_type = 'sender_email'
  AND operator = 'equals'
  AND rule_value = 'contact.center@permatabank.co.id'
  AND target_label = 'Transactions'
  AND priority = 260;

DROP INDEX IF EXISTS user_rules_unique_rule;

ALTER TABLE user_rules
    ALTER COLUMN user_id SET NOT NULL;
