CREATE TABLE "courses" (
	"id"	INTEGER,
	"code"	TEXT NOT NULL,
	"department"	TEXT NOT NULL,
	UNIQUE("code","department"),
	PRIMARY KEY("id" AUTOINCREMENT)
);

CREATE TABLE "sections" (
	"id"	INTEGER,
	"code"	TEXT NOT NULL,
	"term"	TEXT NOT NULL,
	"course_id"	INTEGER NOT NULL,
	PRIMARY KEY("id" AUTOINCREMENT),
	UNIQUE("code","term","course_id"),
	FOREIGN KEY("course_id") REFERENCES "courses"("id")
);

CREATE TABLE "watchers" (
	"id"	INTEGER,
	"email"	TEXT NOT NULL,
	"section_id"	INTEGER NOT NULL,
	UNIQUE("email","section_id"),
	PRIMARY KEY("id" AUTOINCREMENT),
	FOREIGN KEY("section_id") REFERENCES "sections"("id")
);