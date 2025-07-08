import { Coll } from "../../common/api/Coll.ts";
import { by } from "../../common/helpers/by.ts";
import { toArray, toNbr } from "../../common/helpers/cast.ts";
import { isEqual, isObject } from "../../common/helpers/check.ts";
import { createReq, Req } from "../../common/helpers/createReq.ts";
import { toErr } from "../../common/helpers/err.ts";
import {
  Keys,
  ModelBase,
  ModelCreate,
  ModelUpsert,
} from "../../common/api/models.ts";
import { deleteKey } from "../../common/helpers/object.ts";
import { stringify } from "../../common/helpers/json.ts";

export type OdooMethod =
  | "search"
  | "read"
  | "search_read"
  | "create"
  | "write"
  | "unlink";

export type Prop<T> =
  | [string, Keys<T>, (v: any) => any]
  | [string, Keys<T>, (v: any) => any, (v: any) => any];

export type OdooModel<T> = ModelBase & {
  key?: string;
  remote?: Partial<T>;
  deleted?: boolean;
  group?: string;
};

export interface OdooCredential {
  url: string;
  db: string;
  login: string;
  password: string;
}

export class Odoo {
  req: Req;
  userId: any;

  constructor(public credential: OdooCredential, public group: string) {
    this.req = createReq({ baseUrl: credential.url });
  }

  async login() {
    const { db, login, password } = this.credential;
    console.debug("odoo login", login);

    const auth = await this.req("POST", `web/session/authenticate`, {
      json: { jsonrpc: "2.0", params: { db, login, password } },
    });
    if (!auth) throw new Error("error odoo login");

    this.userId = auth.result.uid;
  }

  call(model: string, method: OdooMethod, ...args: any[]) {
    const userId = this.userId;
    const { db, password } = this.credential;
    const json = {
      jsonrpc: "2.0",
      method: "call",
      params: {
        service: "object",
        method: "execute_kw",
        args: [
          db,
          userId,
          password,
          model,
          method,
          ...args,
        ],
      },
    };
    return this.req<any>("POST", `jsonrpc`, { json }).then((data) => {
      const error = data.error;
      if (error) {
        const { code, message, data } = error;
        const args = data.arguments;
        throw toErr(`${message} (${code}) : ${args ? args[0] : null}`);
      }
      return data;
    });
  }

  read(model: string, key: string, fields: any[]) {
    return this.call(model, "read", [[toNbr(key)]], { fields });
  }

  list(model: string, fields: any) {
    return this.call(model, "search_read", [], { fields, limit: 99999 });
  }

  create(model: string, values: any) {
    return this.call(model, "create", [values]);
  }

  update(model: string, key: string, changes: any) {
    return this.call(model, "write", [[toNbr(key)], changes]);
  }

  delete(model: string, key: string) {
    return this.call(model, "unlink", [[toNbr(key)]]);
  }

  coll<T extends OdooModel<T>>(
    model: string,
    props: Prop<T>[],
    pbColl: Coll<T>,
  ) {
    return new OdooColl<T>(this, model, props, pbColl, this.group);
  }
}

class OdooColl<T extends OdooModel<T>> {
  fields: string[];
  select: Keys<T>[];

  constructor(
    public odoo: Odoo,
    public model: string,
    public props: Prop<T>[],
    public pbColl: Coll<T>,
    public group: string,
  ) {
    this.fields = [
      "id",
      ...props.map((p) => p[0]),
    ];
    this.select = [
      "id" as Keys<T>,
      "key" as Keys<T>,
      "remote" as Keys<T>,
      "deleted" as Keys<T>,
      ...props.map((p) => p[1]),
    ];
  }

  dataToRemote(data: any): Partial<T> {
    const remote: any = {};
    for (const prop of this.props) {
      const [dataProp, remoteProp, toRemote] = prop;
      remote[remoteProp] = toRemote(data[dataProp]);
    }
    remote.key = data.id;
    // console.info('dataToRemote', data, remote);
    return remote;
  }

  remoteToData(remote: Partial<T>) {
    const data: any = {};
    for (const prop of this.props) {
      const [dataProp, remoteProp, toRemote, toData] = prop;
      data[dataProp] = (toData || toRemote)(remote[remoteProp]);
    }
    return data;
  }

  async list() {
    const data = await this.odoo.list(this.model, this.fields);
    // console.info('list', data);
    const items = toArray(data.result)
      .map((data) => this.dataToRemote(data))
      .filter(isObject);
    return items;
  }

  async get(key: string) {
    const data = await this.odoo.read(this.model, key, this.fields);
    return this.dataToRemote(data.result[0]);
  }

  async create(item: T) {
    console.info("create", item);
    const data = this.remoteToData(item);
    console.info("create data", data);
    const response = await this.odoo.create(this.model, data);
    const key = response.result;
    if (!key || typeof key !== "string") {
      throw toErr("no odoo create : " + stringify(response));
    }

    const remote = await this.get(key);
    console.info("create remote", item.id, key, remote);

    await this.pbColl.update(
      item.id,
      { remote: deleteKey({ ...remote }, "key") } as Partial<ModelUpsert<T>>,
    );
  }

  async update(item: T) {
    const key = item.key;
    if (!key) throw toErr("no key");
    console.debug("update", item.id, key, item);
    const data = this.remoteToData(item);
    console.debug("update data", key, data);

    const response = await this.odoo.update(this.model, key, data);
    console.debug("update response", response);
    if (response.result !== true) {
      throw toErr("no odoo update : " + stringify(response));
    }

    const remote = await this.get(key);
    console.debug("update remote", item.id, key, remote);

    await this.pbColl.update(item.id, {
      ...remote,
      remote: deleteKey({ ...remote }, "key"),
    } as Partial<ModelUpsert<T>>);
  }

  async delete(item: T) {
    console.info("delete", item);
    const key = item.key;
    if (!key) throw toErr("no key");

    const response = await this.odoo.delete(this.model, key);
    if (response.result !== true) {
      throw toErr("no odoo delete : " + stringify(response));
    }

    await this.pbColl.delete(item.id);
  }

  async syncItem(
    key: string,
    item: T | null,
    remote: Partial<T> | null,
  ) {
    if (!key) throw toErr("no key");
    if (!item && !remote) throw toErr("no item and no remote");

    const group = this.group;

    if (!item) {
      console.info("sync item no item -> create", remote);
      await this.pbColl.create({
        ...remote,
        remote: deleteKey({ ...remote }, "key"),
        group,
      } as ModelCreate<T>);
      return;
    }

    if (!remote) {
      if (item.remote) {
        console.info("sync item no remote -> delete item", item.id, item.key);
        await this.pbColl.delete(item.id);
        return;
      }

      console.info("sync item no remote -> create remote", item.id);
      await this.create(item);
      return;
    }

    if (item.deleted) {
      console.info("sync item deleted -> delete", item.id, remote.id);
      await this.delete(item);
      return;
    }

    const itemRemote = item.remote;
    if (!itemRemote) {
      if (remote) {
        await this.pbColl.update(item.id, {
          ...remote,
          remote: deleteKey({ ...remote }, "key"),
          group,
        } as ModelCreate<T>);
        return;
      }
      return;
    }

    const remoteUpdate: Partial<T> = {};
    for (const prop of this.props) {
      const [_dataProp, remoteProp] = prop;
      // On regarde si la prop de l'object source à changé
      if (!isEqual(itemRemote[remoteProp], remote[remoteProp])) {
        remoteUpdate[remoteProp] = remote[remoteProp];
      }
    }

    // S'il y a modification on synchronise les propriétés
    if (Object.keys(remoteUpdate).length > 0) {
      console.info("sync item update", item, remote, remoteUpdate);
      await this.pbColl.update(item.id, {
        ...remote,
        remote: deleteKey({ ...remote }, "key"),
        group,
      });
      return;
    }

    const itemUpdate: Partial<T> = {};
    for (const prop of this.props) {
      const [_dataProp, remoteProp] = prop;
      if (!isEqual(item[remoteProp], remote[remoteProp])) {
        itemUpdate[remoteProp] = item[remoteProp];
      }
    }

    if (Object.keys(itemUpdate).length > 0) {
      console.info("sync remote update", item, remote, itemUpdate);
      await this.update(item);
      return;
    }

    return;
  }

  async sync() {
    console.info("sync", this.model);

    const remotes = await this.list();
    console.debug("remotes length", remotes.length);

    const items = await this.pbColl.find({ group: this.group }, {
      select: this.select,
    });
    console.debug("items length", items.length);

    const itemByKey = by(items, (i) => i.key);
    const remoteByKey = by(remotes, (r) => r.key);

    const keys = Object.keys({ ...itemByKey, ...remoteByKey });

    for (const key of keys) {
      const item = itemByKey[key];
      const remote = remoteByKey[key];

      await this.syncItem(key, item, remote).catch((err) => {
        console.debug("error sync item", item, remote, toErr(err).message);
      });
    }
  }
}
