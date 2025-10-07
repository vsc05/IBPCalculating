package ds

type CalcUPS struct {
	ID uint `gorm:"primaryKey;autoIncrement"`

	BidID       uint `gorm:"not null"; json:"bidId"`
	ComponentID uint `gorm:"not null"; json:"componentId"`

	BatteryLife     int `gorm:"type:int;not null;default:0"; json:"battery_life"`
	IncomingCurrent int `gorm:"type:int;not null;default:0"; json:"incoming_power"`
	CalculatedPower int `gorm:"type:int;"`

	Bid       BidUPS    `gorm:"foreignKey:BidID"`
	Component Component `gorm:"foreignKey:ComponentID"`
}

// -- Adminer 5.4.0 PostgreSQL 17.6 dump

// DROP TABLE IF EXISTS "bid_components";
// DROP SEQUENCE IF EXISTS bid_components_id_seq;
// CREATE SEQUENCE bid_components_id_seq INCREMENT 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1;

// CREATE TABLE "public"."bid_components" (
//     "id" bigint DEFAULT nextval('bid_components_id_seq') NOT NULL,
//     "bid_id" bigint NOT NULL,
//     "component_id" bigint NOT NULL,
//     "battery_life" bigint NOT NULL,
//     "incoming_current" bigint,
//     "calculated_power" bigint,
//     "is_delete" boolean DEFAULT false,
//     CONSTRAINT "bid_components_pkey" PRIMARY KEY ("id")
// )
// WITH (oids = false);

// CREATE UNIQUE INDEX idx_bid_component ON public.bid_components USING btree (bid_id, component_id);

// INSERT INTO "bid_components" ("id", "bid_id", "component_id", "battery_life", "incoming_current", "calculated_power", "is_delete") VALUES
// (3,	2,	3,	100,	220,	NULL,	'0'),
// (4,	2,	1,	100,	220,	NULL,	'0'),
// (10,	6,	3,	100,	220,	NULL,	'0'),
// (1,	1,	1,	1,	0,	0,	'0'),
// (2,	1,	2,	1,	0,	0,	'0'),
// (9,	5,	1,	10,	100,	1700,	'0'),
// (5,	3,	3,	10,	10,	7100,	'0'),
// (7,	3,	2,	10,	10,	7330,	'0'),
// (11,	7,	1,	20,	20,	NULL,	'0'),
// (12,	7,	2,	20,	20,	NULL,	'0'),
// (13,	7,	3,	23,	23,	NULL,	'0'),
// (8,	4,	2,	100,	220,	NULL,	'0');

// ALTER TABLE ONLY "public"."bid_components" ADD CONSTRAINT "fk_bid_components_component" FOREIGN KEY (component_id) REFERENCES components(id) NOT DEFERRABLE;
// ALTER TABLE ONLY "public"."bid_components" ADD CONSTRAINT "fk_bids_components" FOREIGN KEY (bid_id) REFERENCES bids(id) NOT DEFERRABLE;

// -- 2025-09-23 22:04:09 UTC
