import { Node } from 'node-red';
import PocketBase, { ClientResponseError } from 'pocketbase';

export interface PBAuth {
    url: string;
    authCollection: string;
    username: string;
    password: string;
    token: string;
}

export const isArray = (a: unknown): boolean => Array.isArray(a);
export const isObject = (a: unknown): boolean => typeof a === 'object';
export const isString = (a: unknown): boolean => typeof a === 'string';

export const isPBAuth = (a: unknown): boolean => (
    isObject(a) &&
    isString((a as PBAuth).token) &&
    isString((a as PBAuth).url) &&
    isString((a as PBAuth).authCollection) &&
    isString((a as PBAuth).username) &&
    isString((a as PBAuth).password)
);

export const isPBAuthEquals = (a: PBAuth, b: PBAuth): boolean => (
    a && b &&
    a.url === b.url &&
    a.authCollection === b.authCollection &&
    a.username === b.username &&
    a.password === b.password
);

export const pbAuthInfo = (node: Node, msgAuth: Partial<PBAuth> = {}): PBAuth => {
    if (isPBAuth(msgAuth)) {
        return msgAuth as PBAuth;
    }

    const ctx = node.context();
    const flowAuth = ctx.flow.get('pbAuth') as PBAuth|undefined;
    if (isPBAuth(flowAuth) && isPBAuthEquals(flowAuth as PBAuth, msgAuth as PBAuth)) {
        return flowAuth as PBAuth;
    }

    const env = process.env;
    const url = msgAuth.url || env.PB_URL || '';
    const authCollection = msgAuth.authCollection || env.PB_AUTH_COLLECTION || '_superusers';
    const username = msgAuth.username || env.PB_USERNAME || 'admin';
    const password = msgAuth.password || env.PB_PASSWORD || '';
    const token = '';

    const newAuth: PBAuth = { url, authCollection, username, password, token };
    ctx.flow.set('pbAuth', newAuth);

    return newAuth;
}

export const requiredError = (name: string) => {
    const msg = `${name} is required`;
    return new Error(msg);
}

export const pbAuth = async (node: Node, auth: PBAuth): Promise<{ pb: PocketBase, auth: PBAuth }> => {
    const { url, authCollection, username, password } = auth;
    let { token } = auth;

    if (!url) throw requiredError('PB Url');
    if (!authCollection) throw requiredError('PB Auth Collection');
    if (!username) throw requiredError('PB Username');
    if (!password) throw requiredError('PB Password');

    const ctx = node.context();
    const pb = new PocketBase(url);

    if (token) {
        pb.authStore.save(token);
        try {
            const authData = await pb.collection(authCollection).authRefresh();
            token = authData.token;
        } catch (error) {
            node.debug(`PB token expired or invalid ${error}`);
            token = '';
        }
    }

    if (!token) {
        node.debug(`PB connecting... "${username}"`);
        try {
            const authData = await pb.collection(authCollection).authWithPassword(username, password);
            token = authData.token;
            if (!isString(token)) throw new Error('no token ???');
            node.debug(`PB connected`);
        } catch (error) {
            const infoMsg = JSON.stringify({
                url,
                authCollection,
                username,
                passwordLength: password.length,
            }, null, 2);
            let errorMsg = String(error);
            if (error instanceof ClientResponseError) {
                const errorJson = JSON.stringify(error.toJSON(), null, 2);
                node.error(`PB Auth failed ${infoMsg} : ${error.status} ${error.url} ${errorJson}`);                
            }
            else {
                node.error(`PB Auth failed ${infoMsg} : ${errorMsg}`);
            }
            throw error;
        }
    }

    const newAuth = { ...auth, token };
    ctx.flow.set('pbAuth', newAuth);

    return { pb, auth: newAuth };
}

/** Get authenticated PocketBase client from msg or auto-authenticate */
export const pbAutoAuth = async (node: Node, msg: any): Promise<PocketBase> => {
    if (msg.pb instanceof PocketBase) {
        return msg.pb;
    }
    const { auth, pb } = await pbAuth(node, pbAuthInfo(node, msg.pbAuth));
    msg.pbAuth = auth;
    msg.pb = pb;
    return pb;
}

