-- FitCheck schema: profiles, items, trips, outfits, wear events

-- Style profile (1:1 with auth.users)
create table if not exists profiles (
  id uuid primary key references auth.users(id) on delete cascade,
  display_name text,
  style_primary text default 'Casual',
  style_secondary text[] default '{}',
  likes jsonb default '{}',
  dislikes jsonb default '{}',
  comfort_bias int default 7 check (comfort_bias between 1 and 10),
  photo_look_bias int default 5 check (photo_look_bias between 1 and 10),
  no_repeat_top_days int default 3,
  body_notes text,
  default_formality int default 5 check (default_formality between 1 and 10),
  created_at timestamptz default now(),
  updated_at timestamptz default now()
);

-- Clothing items
create type item_category as enum (
  'tshirt', 'shirt', 'pants', 'shorts', 'jacket', 'shoes', 'accessory'
);

create type item_status as enum ('active', 'dirty', 'packed', 'retired');

create table if not exists items (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null references auth.users(id) on delete cascade,
  storage_path text not null,
  category item_category not null,
  subcategory text,
  name text,
  main_color text,
  secondary_colors text[] default '{}',
  pattern text,
  material text,
  fit text,
  formality int default 5 check (formality between 1 and 10),
  season_tags text[] default '{}',
  rain_ok boolean default false,
  activity_tags text[] default '{}',
  vibe_tags text[] default '{}',
  pair_hints jsonb default '{}',
  ai_raw jsonb,
  user_corrected boolean default false,
  status item_status default 'active',
  wear_count int default 0,
  last_worn_at timestamptz,
  created_at timestamptz default now(),
  updated_at timestamptz default now()
);

create index items_user_id_idx on items(user_id);
create index items_category_idx on items(user_id, category);

-- Trips
create table if not exists trips (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null references auth.users(id) on delete cascade,
  location text not null,
  lat double precision,
  lng double precision,
  start_date date not null,
  end_date date not null,
  activities text[] default '{}',
  formality text default 'casual',
  look_goal text default 'balanced',
  laundry boolean default false,
  luggage text default 'carry_on',
  created_at timestamptz default now()
);

-- Weather cache
create table if not exists weather_cache (
  id uuid primary key default gen_random_uuid(),
  lat double precision not null,
  lng double precision not null,
  date date not null,
  temp_high double precision,
  temp_low double precision,
  precip_probability double precision,
  wind_speed double precision,
  fetched_at timestamptz default now(),
  unique(lat, lng, date)
);

-- Outfit sets
create table if not exists outfit_sets (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null references auth.users(id) on delete cascade,
  trip_id uuid references trips(id) on delete cascade,
  plan_kind text default 'today',
  day_index int,
  label text,
  item_ids uuid[] not null default '{}',
  why text,
  score numeric(3,1),
  variant text default 'base',
  created_at timestamptz default now()
);

-- Wear events (outfit memory)
create table if not exists wear_events (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null references auth.users(id) on delete cascade,
  outfit_set_id uuid references outfit_sets(id),
  item_ids uuid[] not null default '{}',
  worn_at timestamptz default now(),
  source text default 'recommended',
  accepted boolean default true,
  user_rating int check (user_rating between 1 and 5)
);

-- Fit checks
create table if not exists fit_checks (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null references auth.users(id) on delete cascade,
  photo_path text not null,
  outfit_set_id uuid references outfit_sets(id),
  item_ids uuid[] default '{}',
  score numeric(3,1),
  critique jsonb,
  suggested_swaps jsonb,
  created_at timestamptz default now()
);

-- Row Level Security
alter table profiles enable row level security;
alter table items enable row level security;
alter table trips enable row level security;
alter table outfit_sets enable row level security;
alter table wear_events enable row level security;
alter table fit_checks enable row level security;

create policy "Users manage own profile" on profiles
  for all using (auth.uid() = id);

create policy "Users manage own items" on items
  for all using (auth.uid() = user_id);

create policy "Users manage own trips" on trips
  for all using (auth.uid() = user_id);

create policy "Users manage own outfits" on outfit_sets
  for all using (auth.uid() = user_id);

create policy "Users manage own wear events" on wear_events
  for all using (auth.uid() = user_id);

create policy "Users manage own fit checks" on fit_checks
  for all using (auth.uid() = user_id);

-- Storage buckets (run in Supabase dashboard or via CLI)
-- insert into storage.buckets (id, name, public) values ('closet', 'closet', false);
-- insert into storage.buckets (id, name, public) values ('fit-checks', 'fit-checks', false);
