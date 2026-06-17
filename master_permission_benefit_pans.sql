-- ============================================================================
-- MASTER DATA SEEDER — KAI App
-- Permissions, Benefits, and Plans (master/dictionary only)
-- Role-permission and plan-benefit assignments are done separately via UI.
-- Generated: 2026-06-04
-- ============================================================================
-- Idempotent: safe to run repeatedly (ON CONFLICT).
-- ============================================================================

BEGIN;

-- ============================================================================
-- 1. PERMISSIONS (master)
-- ============================================================================
INSERT INTO permissions (key, display_name, description, scope, category, risk_level) VALUES
  -- General content
  ('read_content',               'Read content',               'Read news, events, directory, and announcements',              'global',    'content',    'low'),
  ('join_community',             'Join community',             'Join a community',                                              'global',    'content',    'low'),
  ('post_content',               'Post content',               'Create posts within a community',                               'community', 'content',    'low'),
  ('moderate_posts',             'Moderate posts',             'Moderate community posts',                                      'community', 'moderation', 'medium'),
  ('manage_members',             'Manage community members',   'Manage members of a community',                                 'community', 'member',     'medium'),
  ('delete_content',             'Delete content',             'Delete community content',                                      'community', 'moderation', 'high'),
  ('create_community',           'Create community',           'Create and manage communities',                                 'global',    'content',    'low'),
  ('create_store',               'Create store',               'Create a store/merchant listing',                               'global',    'content',    'low'),
  ('create_event',               'Create event',               'Create events',                                                 'global',    'content',    'low'),
  ('ask_qna',                    'Ask in Q&A',                 'Ask questions in Q&A',                                          'global',    'content',    'low'),
  ('answer_qna',                 'Answer in Q&A',              'Answer questions in Q&A',                                       'global',    'content',    'low'),
  ('moderate_qna',               'Moderate Q&A',               'Moderate Q&A questions and answers',                            'global',    'moderation', 'medium'),
  ('post_announcement',          'Post announcement',          'Create announcements',                                          'global',    'content',    'medium'),
  ('post_announcement_regional', 'Post regional announcement', 'Create region-level announcements',                             'regional',  'content',    'medium'),
  ('manage_directory',           'Manage directory',           'Manage directory/listings',                                     'global',    'admin',      'medium'),
  ('assign_role',                'Assign role to user',        'Assign roles to users',                                         'global',    'admin',      'high'),
  ('manage_region',              'Manage regional settings',   'Manage region settings',                                        'regional',  'admin',      'high'),
  ('view_analytics',             'View analytics',             'View analytics dashboards',                                     'global',    'admin',      'low'),
  ('manage_users',               'Manage users',               'Manage users in backoffice',                                    'global',    'admin',      'high'),
  ('manage_plans',               'Manage plans',               'Manage subscription plans and benefits',                        'global',    'admin',      'high'),
  ('approve_subscription',       'Approve subscription',       'Approve subscription upgrade requests',                         'global',    'admin',      'medium'),
  -- News module
  ('post_news',                  'Post news article',          'Submit a news article (approval depends on role)',              'global',    'content',    'medium'),
  ('create_news',                'Create news article',        'Create news articles in backoffice',                            'global',    'content',    'medium'),
  ('edit_news',                  'Edit news article',          'Edit news articles, including published ones',                  'global',    'content',    'medium'),
  ('publish_news',               'Publish news article',       'Publish news articles from draft',                              'global',    'content',    'medium'),
  ('approve_news',               'Approve news for publishing','Approve or reject news articles from Pro members',               'global',    'admin',      'high'),
  ('manage_news_category',       'Manage news categories',     'Create, update, and delete news categories',                    'global',    'admin',      'medium'),
  ('manage_news_source',         'Manage news sources',        'Register and edit RSS sources and selectors',                   'global',    'system',     'high'),
  ('manage_news_source_config',  'Configure news source',      'Set schedule, auto-publish, AI cleanup, auto-translate, scrape','global',    'system',     'high'),
  ('manage_news_settings',       'Manage news settings',       'Manage system languages and translation settings',              'global',    'system',     'high'),
  ('moderate_news_comment',      'Moderate news comments',     'Delete news comments for moderation',                           'global',    'moderation', 'medium')
ON CONFLICT (key) DO UPDATE SET
  display_name = EXCLUDED.display_name,
  description  = EXCLUDED.description,
  scope        = EXCLUDED.scope,
  category     = EXCLUDED.category,
  risk_level   = EXCLUDED.risk_level,
  updated_at   = NOW();


-- ============================================================================
-- 2. BENEFIT MASTER
-- ============================================================================
INSERT INTO benefit_master (key, display_name, description) VALUES
  ('read_content',        'Read content',        'Read news, events, directory, and announcements'),
  ('join_community',      'Join community',       'Join a community'),
  ('post_community',      'Post in community',    'Post within communities'),
  ('ask_qna',             'Ask in Q&A',           'Ask questions in Q&A'),
  ('post_news',           'Post news',            'Submit news articles (approval handled by role system)'),
  ('create_community',    'Create community',     'Create and manage communities'),
  ('create_store',        'Create store',         'Create a store/merchant listing'),
  ('create_event',        'Create event',         'Create events'),
  ('view_analytics',      'View analytics',       'View analytics dashboards'),
  ('bookmark_news',       'Bookmark news',        'Save news articles to read later'),
  ('request_translation', 'Request translation',  'Request news article translation into another language'),
  ('ad_free',             'Ad-free experience',   'Browse without advertisements')
ON CONFLICT (key) DO UPDATE SET
  display_name = EXCLUDED.display_name,
  description  = EXCLUDED.description;


-- ============================================================================
-- 3. PLANS
-- ============================================================================
INSERT INTO plans (name, price, currency, duration_days, description, is_default, status) VALUES
  ('Standard', 0,     'IDR', 36500, 'Free default plan with basic read and interaction access',         true,  'active'),
  ('Pro',      50000, 'IDR', 30,    'Paid plan with full access including content creation and news',    false, 'active')
ON CONFLICT (LOWER(name)) DO UPDATE SET
  price         = EXCLUDED.price,
  description   = EXCLUDED.description,
  duration_days = EXCLUDED.duration_days,
  updated_at    = NOW();


COMMIT;

-- ============================================================================
-- VERIFICATION (run manually after seeding)
-- ============================================================================
-- SELECT key, scope, category, risk_level FROM permissions ORDER BY category, key;
-- SELECT key, display_name FROM benefit_master ORDER BY key;
-- SELECT name, price, is_default, status FROM plans ORDER BY price;
