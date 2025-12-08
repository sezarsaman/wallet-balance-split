import http from 'k6/http';
import { sleep, check } from 'k6';
import { randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js'; // برای رندوم

export const options = {
  vus: 100,  // تعداد کاربران مجازی ثابت (۱۰۰ کاربر همزمان)
  duration: '1h',  // مدت تست: ۱ ساعت
  thresholds: {
    http_req_duration: ['p(95)<200'], // ۹۵% درخواست‌ها زیر ۲۰۰ms
    checks: ['rate>0.99'], // ۹۹% چک‌ها موفق
  },
};

export default function () {
  const userId = randomIntBetween(1, 100); // user_id رندوم برای جلوگیری از cache
  const baseUrl = 'http://localhost:8080'; // تارگت اپ داکر

  // ۷۰% charge
  if (Math.random() < 0.6) {
    const payload = JSON.stringify({
      user_id: userId,
      amount: randomIntBetween(1000, 10000),
      idempotency_key: 'unique-' + Math.random(), // کلید یکتا برای جلوگیری از دوبار شارژ
      release_at: randomIsoDateTimeWithin3Hours() // اختیاری، برای تست split balance
    });
    const res = http.post(`${baseUrl}/charge`, payload, { headers: { 'Content-Type': 'application/json' } });
    check(res, { 'charge success': (r) => r.status === 200 });
  }
  // ۲۰% get balance or transactions
  else if (Math.random() < 0.9) {
    if (Math.random() < 0.5) {
      const res = http.get(`${baseUrl}/balance?user_id=${userId}`);
      check(res, { 'balance success': (r) => r.status === 200 });
    } else {
      const res = http.get(`${baseUrl}/transactions?user_id=${userId}&page=1&limit=10`); // صفحه‌بندی
      check(res, { 'transactions success': (r) => r.status === 200 });
    }
  }
  // ۱۰% withdraw (با چک available balance)
  else {
    const payload = JSON.stringify({
      user_id: userId,
      idempotency_key: 'unique-' + Math.random(), // کلید یکتا برای جلوگیری از دوبار شارژ
      amount: randomIntBetween(1000, 5000),
    });
    const res = http.post(`${baseUrl}/withdraw`, payload, { headers: { 'Content-Type': 'application/json' } });
    check(res, { 'withdraw success': (r) => r.status === 200 || r.status === 202 }); // ۲۰۲ برای async
  }

  sleep(1); // wait بین درخواست‌ها برای شبیه‌سازی واقعی
}

function randomIsoDateTimeWithin3Hours() {
    const now = new Date();
    const threeHoursLater = new Date(now.getTime() + 3 * 60 * 60 * 1000);

    // تولید یک timestamp رندوم بین now و threeHoursLater
    const randomTimestamp =
        now.getTime() +
        Math.random() * (threeHoursLater.getTime() - now.getTime());

    return new Date(randomTimestamp).toISOString();
}