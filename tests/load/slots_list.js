import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
    scenarios: {
        slots_list_rps: {
            executor: "constant-arrival-rate",
            rate: 100,
            timeUnit: "1s",
            duration: "60s",
            preAllocatedVUs: 50,
            maxVUs: 150,
        },
    },
    thresholds: {
        http_req_failed: ["rate<0.001"],
        http_req_duration: ["p(95)<200"],
    },
};

const baseURL = __ENV.BASE_URL || "http://localhost:8080";
const roomID = __ENV.ROOM_ID;
const token = __ENV.TOKEN;
const date = __ENV.DATE;

if (!roomID || !token || !date) {
    throw new Error("ROOM_ID, TOKEN and DATE env vars are required");
}

export default function () {
    const res = http.get(`${baseURL}/rooms/${roomID}/slots/list?date=${date}`, {
        headers: {
            Authorization: `Bearer ${token}`,
        },
    });

    check(res, {
        "status is 200": (r) => r.status === 200,
        "body has slots": (r) => r.json("slots") !== undefined,
    });

    sleep(0.1);
}