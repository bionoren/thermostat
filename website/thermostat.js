let zoneID = 0;
let zones = refreshZones();
function refreshZones() {
    return new Promise(function (resolve, reject) {
        request("/v1/zones", "").done(function (data) {
            let zoneArray = JSON.parse(data);
            let zones = new Map();
            zoneArray.forEach(function (item) {
                zones.set(item.ID, item)
            });
            resolve(zones);
        }).fail(function (data) {
            reject(data.responseText)
        });
    });
}
let modes;
function refreshModes() {
    return new Promise(function (resolve, reject) {
        request("/v1/mode", {zoneID: zoneID}).done(function (data) {
            let modeArray = JSON.parse(data);
            let modes = new Map();
            modeArray.forEach(function (item) {
                modes.set(item.ID, item)
            });
            resolve(modes);
        }).fail(function (data) {
            reject(data.responseText)
        });
    });
}
let schedules;
function refreshSchedules() {
    return new Promise(function (resolve, reject) {
        request("/v1/schedule", {zoneID: zoneID}).done(function (data) {
            let scheduleArray = JSON.parse(data);
            let schedules = new Map();
            scheduleArray.forEach(function (item, index) {
                schedules.set(item.ID, item);
            });
            resolve(schedules);
        }).fail(function (data) {
            reject(data.responseText);
        });
    });
}

let lastData = null;
let currentData = null

async function refresh() {
    let z = await zones;
    $("#mainList")[0].innerHTML = "";
    z.forEach(function (zone) {
        $("#mainList")[0].appendChild(tableCell(zone.name, "", "", function () {
            zoneID = zone.ID
            currentData = new Promise(function refresh(resolve, reject) {
                              request("/v1/status", {zoneID: zoneID}).done(function(data) {
                                  resolve(JSON.parse(data));
                              }).fail(function(data) {
                                  reject(data.responseText);
                              })
                          });
            modes = refreshModes()
            schedules = refreshSchedules()
            loadPage("zone.html")
        }));
    });
}

async function refreshZonePage() {
    let data = lastData;
    if(data == null) {
        lastData = await currentData;
        data = lastData;
    }

    $("#temp")[0].innerHTML = roundTemp(data.Temperature);

    let m = await modes;
    $("#modeSelect").empty();
    m.forEach(function (item) {
        let opt = document.createElement("option");
        opt.value = item.ID;
        opt.innerHTML = item.name + " (" + item.minTemp + ", " + item.maxTemp + " ±" + item.correction + ")"; // whatever property it has

        // then append it to the select element
        $("#modeSelect")[0].appendChild(opt);
    });

    $("#zoneSchedule")[0].innerHTML = JSON.stringify(schedules, null, 2);
}

function loadPage(page, args) {
    request("/v1/"+page, "").done(function(data) {
        $("body")[0].innerHTML = data;

        switch(page) {
            case "main.html":
                refresh();
                break;
            case "zone.html":
                refreshZonePage();
                break;
            case "modes.html":
                refreshModesPage();
                break;
            case "editModes.html":
                editMode(args);
                break;
            case "schedules.html":
                refreshSchedulesPage();
                break;
            case "addSchedule.html":
                addSchedule();
                break;
        }
    }).fail(function(data) {
        console.log("error: "+data.responseText)
    })
}

async function editMode(id) {
    let m = await modes;
    let v = m.get(id);
    $("#modeID")[0].value = v.ID;
    $("#name")[0].value = v.name;
    $("#minTemp")[0].value = v.minTemp;
    $("#maxTemp")[0].value = v.maxTemp;
    $("#offset")[0].value = v.correction;
}

async function refreshModesPage() {
    let m = await modes;
    $("#mainList")[0].innerHTML = "";
    m.forEach(function (mode) {
        let subtitle = roundTemp(mode.minTemp)+" to "+roundTemp(mode.maxTemp)+"℉";
        let detail = "±"+roundTemp(mode.correction)+"℉";
        $("#mainList")[0].appendChild(tableCell(mode.name, subtitle, detail, function () {
            loadPage("editModes.html", mode.ID)
        }));
    });
}

function prettyDate(datestring) {
    let d = new Date(datestring);
    return d.toLocaleDateString("en-US")+" "+d.toTimeString().substr(0,5);
}

function secondsToTime(seconds) {
    let hours = seconds / 3600;
    let minutes = seconds % 3600;
    if(hours == 0) {
        hours = "00";
    } else if(hours < 10) {
        hours = "0"+hours;
    }
    if(minutes == 0) {
        minutes = "00";
    } else if(minutes < 10) {
        minutes = "0"+minutes;
    }

    return hours+":"+minutes;
}

function maskToWeekdays(mask) {
    let weekdays = "";
    if(mask & 2) {
        weekdays += "<span class='selectedDay'>S</span>";
    } else {
        weekdays += "<span class='unselectedDay'>S</span>";
    }
    if(mask & 2 << 1) {
        weekdays += "<span class='selectedDay'>M</span>";
    } else {
        weekdays += "<span class='unselectedDay'>S</span>";
    }
    if(mask & 2 << 2) {
        weekdays += "<span class='selectedDay'>T</span>";
    } else {
        weekdays += "<span class='unselectedDay'>S</span>";
    }
    if(mask & 2 << 3) {
        weekdays += "<span class='selectedDay'>W</span>";
    } else {
        weekdays += "<span class='unselectedDay'>S</span>";
    }
    if(mask & 2 << 4) {
        weekdays += "<span class='selectedDay'>T</span>";
    } else {
        weekdays += "<span class='unselectedDay'>S</span>";
    }
    if(mask & 2 << 5) {
        weekdays += "<span class='selectedDay'>F</span>";
    } else {
        weekdays += "<span class='unselectedDay'>S</span>";
    }
    if(mask & 2 << 6) {
        weekdays += "<span class='selectedDay'>S</span>";
    } else {
        weekdays += "<span class='unselectedDay'>S</span>";
    }

    return weekdays;
}

async function refreshSchedulesPage() {
    let s = await schedules;
    let m = await modes;
    $("#mainList")[0].innerHTML = "";
    s.forEach(function (schedule) {
        let subtitle = maskToWeekdays(schedule.dayOfWeek)+"<br>"+secondsToTime(schedule.startTime)+" - "+secondsToTime(schedule.endTime);
        let detail = "start: "+prettyDate(schedule.startDay)+"<br>end: "+prettyDate(schedule.endDay);
        $("#mainList")[0].appendChild(tableCell(m.get(schedule.modeID).name, subtitle, detail, function () {
            loadPage("editSchedule.html", schedule.ID)
        }));
    });
}

async function addSchedule() {
    flatpickr("#startTime", {
        enableTime: true,
        noCalendar: true,
        altInput: true,
        altFormat: "h:i K",
        dateFormat: "H:i"
    });
    flatpickr("#endTime", {
        enableTime: true,
        noCalendar: true,
        altInput: true,
        altFormat: "h:i K",
        dateFormat: "H:i"
    });
    flatpickr("#startDate", {
        defaultDate: "today",
        enableTime: true,
        altInput: true,
        altFormat: "Y-m-d H:i",
        dateFormat: "Z"
    });
    flatpickr("#endDate", {
        defaultDate: "today",
        enableTime: true,
        altInput: true,
        altFormat: "Y-m-d H:i",
        dateFormat: "Z"
    });

    let m = await modes;
    $("#modeSelect").empty();
    let opt = document.createElement("option");
    opt.innerHTML = "";
    $("#modeSelect")[0].appendChild(opt);
    m.forEach(function (item) {
        let opt = document.createElement("option");
        opt.value = item.ID;
        opt.innerHTML = item.name + " (" + item.minTemp + ", " + item.maxTemp + " ±" + item.correction + ")"; // whatever property it has

        // then append it to the select element
        $("#modeSelect")[0].appendChild(opt);
    });
}

function weekdaysToMask(weekdays) {
    let mask = 0;
    weekdays.forEach(function (value, i) {
       mask = mask | (value << i);
    });
    return mask;
}

function secondsFromTime(time) {
    let parts = time.split(" ");
    return parts[0]*60*60 + parts[1]*60;
}

function addScheduleForm() {
    let data = $('#addForm').serializeArray().reduce(function(obj, item) {
        obj[item.name] = item.value;
        return obj;
    }, {});

    let weekdays = [];
    let checkboxes = $(".dayBox");
    for(let i = 0; i < checkboxes.length; i++) {
        let val = 0;
        if(checkboxes[i].value == "on") {
            val = 1;
        }
        weekdays.push(val);
    }

    let req = {
        ZoneID: zoneID,
        ModeID: parseInt(data.modeID),
        Priority: data.priority,
        DayOfWeek: weekdaysToMask(weekdays),
        StartTime: secondsFromTime(data.startTime),
        EndTime: secondsFromTime(data.startTime),
        StartDay: parseInt(data.startDate),
        EndDay: parseInt(data.endDate),
    };

    request("/v1/schedule/add", req).done(function(data) {
        schedules = refreshSchedules();
        loadPage("schedules.html");
    }).fail(function(data) {
        $("#errors")[0].innerHTML = data.responseText;
    });

    return false;
}

function roundTemp(num) {
    return Math.round(num*10)/10;
}

function tableCell(titleText, subtitleText, detailText, onClick) {
    let title = document.createElement("div");
    title.className = "tablecellTitle";
    title.innerHTML = titleText;

    let subtitle = document.createElement("div");
    subtitle.className = "tablecellSubtitle";
    subtitle.innerHTML = subtitleText;

    let detail = document.createElement("div");
    detail.className = "tablecellDetail";
    detail.innerHTML = detailText;

    let rightDiv = document.createElement("div");
    rightDiv.className = "tablecellRight";
    rightDiv.appendChild(subtitle);
    rightDiv.appendChild(detail);

    let cell = document.createElement("div");
    cell.className = "tablecell";
    cell.appendChild(title);
    cell.appendChild(rightDiv);
    cell.onclick = onClick;

    return cell;
}

function addModeForm() {
    let data = $('#addForm').serializeArray().reduce(function(obj, item) {
        obj[item.name] = item.value;
        return obj;
    }, {});

    let req = {
        ZoneID: zoneID,
        Name: data.name,
        MinTemp: parseFloat(data.minTemp),
        MaxTemp: parseFloat(data.maxTemp),
        Correction: parseFloat(data.offset)
    };

    request("/v1/mode/add", req).done(function(data) {
        modes = refreshModes();
        loadPage("modes.html");
    }).fail(function(data) {
        $("#errors")[0].innerHTML = data.responseText;
    });

    return false;
}

function editModeForm() {
    let data = $('#editForm').serializeArray().reduce(function(obj, item) {
        obj[item.name] = item.value;
        return obj;
    }, {});

    let req = {
        ZoneID: zoneID,
        ID: parseInt(data.ID),
        Name: data.name,
        MinTemp: parseFloat(data.minTemp),
        MaxTemp: parseFloat(data.maxTemp),
        Correction: parseFloat(data.offset)
    };

    request("/v1/mode/edit", req).done(function(data) {
        modes = refreshModes();
        loadPage("modes.html");
    }).fail(function(data) {
        $("#errors")[0].innerHTML = data.responseText;
    });

    return false;
}

// onload stuff

request("/v1/thermostat.css", "").done(function(data) {
    let css = document.createElement('style');
    css.language = 'text/css';
    css.appendChild(document.createTextNode(data)); // Support for the rest
    document.getElementsByTagName("head")[0].appendChild(css);
}).fail(function (data) {
    console.log("failed loading css: "+data.responseText)
});

loadPage("main.html");
