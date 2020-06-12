create table zone (
    id integer primary key asc,
    name text,
    unique(name)
);

create table mode (
    id integer primary key asc,
    zoneID integer,
    name text,
    minTemp real,
    maxTemp real,
    correction real,
    foreign key (zoneID) references zone(id) on delete cascade,
    unique(zoneID, name)
);

create table setting (
    id integer primary key asc,
    zoneID integer,
    modeID integer,
    priority integer,
    dayOfWeek integer,
    startDay timestamp,
    endDay timestamp,
    startTime integer,
    endTime integer,
    foreign key (zoneID) references zone(id) on delete cascade,
    foreign key (modeID) references mode(id) on delete cascade
);

insert into zone (name) values ('default');
insert into mode (zoneID, name, minTemp, maxTemp, correction) select id, 'default', 60, 85, 1 from zone where name = 'default';
insert into mode (zoneID, name, minTemp, maxTemp, correction) select id, 'custom', 60, 85, 1 from zone where name = 'default';
insert into setting (zoneID, modeID, priority, dayOfWeek, startDay, endDay, startTime, endTime) select id, zoneID, 1, 254, datetime(0, 'unixepoch'), datetime(106751991167, 'unixepoch'), 0, 86400 from mode where name = 'default';
