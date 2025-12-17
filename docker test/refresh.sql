-- Adminer 5.4.1 PostgreSQL 18.1 dump

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
    "incoming_current" bigint DEFAULT '0',
    CONSTRAINT "bid_ups_pkey" PRIMARY KEY ("id")
)
WITH (oids = false);

INSERT INTO "bid_ups" ("id", "status", "date_update", "date_finish", "creator_id", "moderator_id", "incoming_current") VALUES
(5,	'отклонен',	'2025-10-20 02:22:22.15148+00',	'2025-10-20 02:22:22.15148+00',	1,	3,	NULL),
(2,	'удалён',	'2025-10-20 02:22:22.15148+00',	'2025-10-20 02:22:22.15148+00',	1,	NULL,	NULL),
(3,	'завершена',	'2025-10-20 02:22:22.15148+00',	'2025-10-21 10:03:20.432851+00',	1,	3,	NULL),
(4,	'рассчитан',	'2025-10-20 02:22:22.15148+00',	'2025-10-27 02:22:22.15148+00',	1,	NULL,	NULL),
(18,	'удален',	'2025-12-03 01:05:37.922798+00',	'2025-12-03 01:13:43.815505+00',	1,	3,	220),
(43,	'удален',	'2025-12-15 02:54:58.793983+00',	'2025-12-15 02:55:04.57527+00',	1,	3,	220),
(19,	'удален',	'2025-12-03 01:17:59.591861+00',	'2025-12-03 01:18:23.151808+00',	5,	3,	220),
(20,	'удален',	'2025-12-03 02:47:17.926705+00',	'2025-12-03 02:47:34.29037+00',	5,	3,	220),
(46,	'сформирован',	'2025-12-15 14:42:34.86644+00',	NULL,	1,	NULL,	228),
(44,	'завершена',	'2025-12-15 14:31:17.531259+00',	'2025-12-15 19:10:47.713128+00',	1,	3,	444),
(45,	'завершена',	'2025-12-15 14:32:19.746883+00',	'2025-12-15 19:16:47.944087+00',	1,	3,	230),
(27,	'завершена',	'2025-12-14 23:13:30.960072+00',	'2025-12-14 23:53:16.714985+00',	1,	3,	220),
(29,	'сформирован',	'2025-12-15 00:46:36.954958+00',	NULL,	1,	NULL,	220),
(30,	'удален',	'2025-12-15 00:49:24.593021+00',	'2025-12-15 00:53:19.638676+00',	1,	3,	220);

DROP TABLE IF EXISTS "calc_ups";
DROP SEQUENCE IF EXISTS calc_ups_id_seq;
CREATE SEQUENCE calc_ups_id_seq INCREMENT 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1;

CREATE TABLE "public"."calc_ups" (
    "id" bigint DEFAULT nextval('calc_ups_id_seq') NOT NULL,
    "bid_id" bigint NOT NULL,
    "component_id" bigint NOT NULL,
    "calculated_power" bigint,
    "battery_life" bigint DEFAULT '0',
    "count" bigint,
    CONSTRAINT "calc_ups_pkey" PRIMARY KEY ("id")
)
WITH (oids = false);

INSERT INTO "calc_ups" ("id", "bid_id", "component_id", "calculated_power", "battery_life", "count") VALUES
(114,	43,	6,	0,	0,	1),
(115,	43,	8,	0,	0,	1),
(116,	43,	4,	0,	0,	1),
(14,	5,	4,	0,	0,	NULL),
(13,	4,	8,	0,	0,	NULL),
(121,	46,	6,	0,	20,	2),
(80,	27,	6,	100,	10,	NULL),
(81,	27,	8,	100,	10,	NULL),
(122,	46,	4,	0,	40,	3),
(118,	44,	8,	90,	30,	15),
(117,	44,	6,	99,	200,	2),
(84,	29,	6,	0,	0,	1),
(85,	29,	4,	0,	0,	1),
(119,	45,	6,	43000,	100,	2),
(15,	3,	6,	800,	0,	NULL),
(12,	3,	4,	1705,	0,	NULL),
(120,	45,	8,	15200,	40,	1),
(17,	2,	6,	0,	0,	NULL);

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
(8,	'http://localhost:9000/test/Pc.png',	'Компьютер',	150,	1.0,	'0'),
(6,	'http://localhost:9000/test/Router.png',	'Rоутер',	200,	1.0,	'0'),
(10,	'http://localhost:9000/test/component_10_1763939867.png',	'string',	100,	0,	'1');

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
(3,	'user3',	'123',	'1'),
(4,	'Djabrail',	'123',	'0'),
(5,	'Karolishkins',	'123',	'0'),
(6,	'user11',	'123456',	'0'),
(2,	'User2',	'123',	'0'),
(1,	'user1',	'123',	'0');

ALTER TABLE ONLY "public"."bid_ups" ADD CONSTRAINT "fk_bid_ups_creator" FOREIGN KEY (creator_id) REFERENCES users(id) NOT DEFERRABLE;
ALTER TABLE ONLY "public"."bid_ups" ADD CONSTRAINT "fk_bid_ups_moderator" FOREIGN KEY (moderator_id) REFERENCES users(id) NOT DEFERRABLE;

ALTER TABLE ONLY "public"."calc_ups" ADD CONSTRAINT "fk_bid_ups_components" FOREIGN KEY (bid_id) REFERENCES bid_ups(id) NOT DEFERRABLE;
ALTER TABLE ONLY "public"."calc_ups" ADD CONSTRAINT "fk_calc_ups_component" FOREIGN KEY (component_id) REFERENCES components(id) NOT DEFERRABLE;

-- 2025-12-15 19:30:31 UTC