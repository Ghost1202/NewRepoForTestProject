import axios from "axios";

const baseUrl = process.env.SSO_BASE_URL ?? "http://localhost:8080";

const now = Date.now();
const email = process.env.SSO_EMAIL ?? `test_${now}@example.com`;
const password = process.env.SSO_PASSWORD ?? "Test12345!";
const login = process.env.SSO_LOGIN ?? `test_${now}`;
const name = process.env.SSO_NAME ?? "Test";
const lastName = process.env.SSO_LAST_NAME ?? "User";

const phone = process.env.SSO_PHONE;
const phonePassword = process.env.SSO_PHONE_PASSWORD;

const client = axios.create({
  baseURL: baseUrl,
  headers: { "Content-Type": "application/json" },
  validateStatus: () => true,
});

const logResult = (label, response) => {
  const status = response?.status ?? "ERR";
  const data = response?.data ?? "";
  console.log(`${label}: ${status}`);
  if (typeof data === "string") {
    console.log(data);
  } else if (data && typeof data === "object") {
    console.log(JSON.stringify(data, null, 2));
  }
};

const run = async () => {
  console.log(`SSO base URL: ${baseUrl}`);
  console.log(`Email: ${email}`);
  console.log(`Login: ${login}`);

  const registerPayload = {
    name,
    last_name: lastName,
    login,
    email,
    password,
  };

  const registerRes = await client.post("/users/register", registerPayload);
  logResult("Register", registerRes);

  const emailLoginRes = await client.post("/users/auth", {
    is_email: true,
    identifier: email,
    password,
  });
  logResult("Login (email)", emailLoginRes);

  const token = emailLoginRes?.data?.token ?? emailLoginRes?.data?.access_token;
  if (token) {
    const profileRes = await client.get("/users/login", {
      headers: { Authorization: `Bearer ${token}` },
    });
    logResult("Profile", profileRes);
  }

  if (phone && phonePassword) {
    const phoneLoginRes = await client.post("/users/auth/phone", {
      phone,
      password: phonePassword,
    });
    logResult("Login (phone)", phoneLoginRes);
  } else {
    console.log("Phone login skipped (set SSO_PHONE and SSO_PHONE_PASSWORD).");
  }
};

run().catch((error) => {
  console.error("SSO smoke test failed:", error?.message ?? error);
  process.exit(1);
});
