-- Adminer 5.4.1 PostgreSQL 18.0 dump

DROP TABLE IF EXISTS "bid_ups";
DROP SEQUENCE IF EXISTS bid_ups_id_seq;
CREATE SEQUENCE bid_ups_id_seq INCREMENT 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1;

CREATE TABLE "public"."bid_ups" (
    "id" bigint DEFAULT nextval('bid_ups_id_seq') NOT NULL,
    "status" text,
    "date_update" timestamptz,
    "date_finish" timestamptz,
    "creator_id" bigint NOT NULL,
    "moderator_id" bigint,
    CONSTRAINT "bid_ups_pkey" PRIMARY KEY ("id")
)
WITH (oids = false);

INSERT INTO "bid_ups" ("id", "status", "date_update", "date_finish", "creator_id", "moderator_id") VALUES
(1,	'черновик',	'2025-10-20 02:22:22.15148+00',	NULL,	1,	NULL),
(2,	'удален',	'2025-10-20 02:22:22.15148+00',	'2025-10-20 02:22:22.15148+00',	1,	NULL),
(3,	'сформирован',	'2025-10-20 02:22:22.15148+00',	'2025-10-20 02:22:22.15148+00',	1,	3),
(4,	'рассчитан',	'2025-10-20 02:22:22.15148+00',	'2025-10-20 02:22:22.15148+00',	1,	NULL),
(5,	'отклонен',	'2025-10-20 02:22:22.15148+00',	'2025-10-20 02:22:22.15148+00',	1,	3);

DROP TABLE IF EXISTS "calc_ups";
DROP SEQUENCE IF EXISTS calc_ups_id_seq;
CREATE SEQUENCE calc_ups_id_seq INCREMENT 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1;

CREATE TABLE "public"."calc_ups" (
    "id" bigint DEFAULT nextval('calc_ups_id_seq') NOT NULL,
    "bid_id" bigint NOT NULL,
    "component_id" bigint NOT NULL,
    "battery_life" bigint DEFAULT '0' NOT NULL,
    "incoming_current" bigint DEFAULT '0' NOT NULL,
    "calculated_power" bigint,
    CONSTRAINT "calc_ups_pkey" PRIMARY KEY ("id")
)
WITH (oids = false);

INSERT INTO "calc_ups" ("id", "bid_id", "component_id", "battery_life", "incoming_current", "calculated_power") VALUES
(8,	1,	4,	100,	50,	5000),
(9,	1,	6,	120,	60,	7200),
(10,	1,	8,	90,	45,	4050),
(11,	2,	6,	80,	40,	3200),
(12,	3,	4,	110,	55,	6050),
(13,	4,	8,	95,	48,	4560),
(14,	5,	4,	105,	52,	5460);

DROP TABLE IF EXISTS "components";
DROP SEQUENCE IF EXISTS components_id_seq;
CREATE SEQUENCE components_id_seq INCREMENT 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1;

CREATE TABLE "public"."components" (
    "id" bigint DEFAULT nextval('components_id_seq') NOT NULL,
    "image" text,
    "title" text,
    "power" bigint,
    "coeff" numeric,
    "is_delete" boolean,
    CONSTRAINT "components_pkey" PRIMARY KEY ("id")
)
WITH (oids = false);

INSERT INTO "components" ("id", "image", "title", "power", "coeff", "is_delete") VALUES
(4,	'http://localhost:9000/test/Monitor.png',	'Монитор',	100,	1.0,	'0'),
(6,	'http://localhost:9000/test/Router.png',	'Роутер',	200,	1.0,	'0'),
(8,	'http://localhost:9000/test/Pc.png',	'Компьютер',	150,	1.0,	'0');

DROP TABLE IF EXISTS "users";
DROP SEQUENCE IF EXISTS users_id_seq;
CREATE SEQUENCE users_id_seq INCREMENT 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1;

CREATE TABLE "public"."users" (
    "id" bigint DEFAULT nextval('users_id_seq') NOT NULL,
    "login" character varying(25) NOT NULL,
    "password" character varying(100) NOT NULL,
    "is_moderator" boolean DEFAULT false,
    CONSTRAINT "users_pkey" PRIMARY KEY ("id")
)
WITH (oids = false);

CREATE UNIQUE INDEX uni_users_login ON public.users USING btree (login);

INSERT INTO "users" ("id", "login", "password", "is_moderator") VALUES
(1,	'user1',	'123',	'0'),
(2,	'user2',	'123',	'0'),
(3,	'user3',	'123',	'1');

ALTER TABLE ONLY "public"."bid_ups" ADD CONSTRAINT "fk_bid_ups_creator" FOREIGN KEY (creator_id) REFERENCES users(id) NOT DEFERRABLE;
ALTER TABLE ONLY "public"."bid_ups" ADD CONSTRAINT "fk_bid_ups_moderator" FOREIGN KEY (moderator_id) REFERENCES users(id) NOT DEFERRABLE;

ALTER TABLE ONLY "public"."calc_ups" ADD CONSTRAINT "fk_bid_ups_components" FOREIGN KEY (bid_id) REFERENCES bid_ups(id) NOT DEFERRABLE;
ALTER TABLE ONLY "public"."calc_ups" ADD CONSTRAINT "fk_calc_ups_component" FOREIGN KEY (component_id) REFERENCES components(id) NOT DEFERRABLE;

-- 2025-10-20 02:26:51 UTC