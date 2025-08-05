export const pbBaseUrl = "http://localhost:8090";

let _token = '';

export const pbFetch = async (info: RequestInit & { url: string; json?: any }) => {
  let { method, url, headers, json, body, ...rest } = info;

  if (json) {
    body = JSON.stringify(json);
    headers = { ...headers, "Content-Type": "application/json" };
  }

  if (_token) headers = { ...headers, "Authorization": `Bearer ${_token}` };

  if (url.startsWith('/')) url = `${pbBaseUrl}${url}`;

  const r = await fetch(url, {
    method,
    headers,
    body,
    ...rest,
  });

  if (!r.ok) throw new Error(`❌ ${method} ${url}: ${r.status}`);

  const data = await r.json();
  return data;
};

export const pbAuth = async () => {
  // Tester différents endpoints possibles
  const adminEmail = Deno.env.get("ADMIN_EMAIL");
  const adminPassword = Deno.env.get("ADMIN_PASSWORD");

  const data = await pbFetch({
    method: "POST",
    url: '/api/collections/_superusers/auth-with-password',
    json: {
      identity: adminEmail,
      password: adminPassword,
    },
  });

  return _token = data.token;
};