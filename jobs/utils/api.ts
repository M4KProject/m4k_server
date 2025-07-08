import { fun } from "../../common/api/functions.ts";
import { apiUrl$, auth$ } from "../../common/api/messages.ts";

export const adminLogin = async () => {
    const identity = Deno.env.get('ADMIN_EMAIL');
    const password = Deno.env.get('ADMIN_PASSWORD');

    apiUrl$.set('http://0.0.0.0:8090/api/');

    const result = await fun('POST', 'collections/_superusers/auth-with-password', {
        form: { identity, password }
    });

    const auth = { ...result.record, token: result.token };
    console.debug("auth id", auth.id);

    auth$.set(auth);

    return auth;
}


// fetch('/api/admins/auth-with-password')

// import { auth$, apiUrl$, groupId$ } from "../../common/api/index.ts";

