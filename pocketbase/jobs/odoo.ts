import { categoryColl, contentColl, productColl } from "../common/api/collections.ts";
import { _ProductModel } from "../common/api/models.ts";
import { toNbr, toStr } from "../common/helpers/cast.ts";
import { toErr } from "../common/helpers/err.ts";
import { adminLogin } from "./utils/api.ts";
import { initConsole, job, setProgress } from "./utils/base.ts";
import { Odoo, OdooCredential } from "./utils/Odoo.ts";

const group = job.group;
if (!group) throw toErr('no group');
console.info("group", group);

initConsole();

await adminLogin();

const odooCredential: OdooCredential = await contentColl.findOne({
  type: "odoo",
  group,
  data: ["!=", null],
}).then((c) => c?.data);
if (!odooCredential) throw new Error("no odoo credential");

const odoo = new Odoo(odooCredential, group);

const odooTextToStr = (v: any) => v === false ? '' : toStr(v);
const strToOdooText = (v: any) => v === '' ? false : toStr(v);

const catIdByKey: Record<string|number, string> = {};
const catKeyById: Record<string, number> = {};

const toCatId = (key: number) => {
  const id = catIdByKey[key];
  console.debug('toCatId', key, '->', id);
  return id;
}
const toCatKey = (id: number) => {
  const key = catKeyById[id];
  console.debug('toCatKey', id, '->', key);
  return id;
}

const odooCategoryColl = odoo.coll("product.category", [
  ["name", "name", toStr],
  ["parent_id", "parent", toCatId, toCatKey],
], categoryColl);

const odooProductColl = odoo.coll("product.product", [
  ["name", "name", toStr],
  ["description", "desc", odooTextToStr, strToOdooText],
  ["active", "disabled", v => !v, v => !v],
  ["list_price", "price", toNbr], // v => v / 100, v => round(v * 100)
  ["nbr_reordering_rules", "order", toNbr],
  ["categ_id", "category", toCatId, toCatKey],
], productColl);

await odoo.login();

const loadCat = async () => {
  const categories = await categoryColl.find({ group }, { select: ['id', 'key'] });
  for (const { id, key } of categories) {
    const k = toNbr(key);
    if (!key || !k) continue;
    catIdByKey[key] = id;
    catIdByKey[k] = id;
    catKeyById[id] = k;
  }
}

// await categoryColl.update({ group, key:"2" }, { name: "Expenses" });
// await categoryColl.update({ group, key:"3" }, { name: "Services" });
// await categoryColl.update({ group, key:"33" }, { name: "Sushi" });
// await categoryColl.update({ group, key:"34" }, { name: "Box Sushis" });
// await categoryColl.update({ group, key:"35" }, { name: "Plateaux Sushis" });
// await categoryColl.update({ group, key:"36" }, { name: "Sauces" });
// await categoryColl.update({ group, key:"37" }, { name: "Sandwich" });
// await categoryColl.update({ group, key:"38" }, { name: "Sandwichs" });
// await categoryColl.update({ group, key:"39" }, { name: "Ingrédients Sandwich" });

setProgress(10);

await loadCat();
setProgress(15);

await odooCategoryColl.sync();
setProgress(30);

await loadCat();
setProgress(35);

await odooProductColl.sync();
setProgress(50);














// const syncItemToRemote = async <T extends OdooModel>(model: string, coll: Coll<T>, props: Prop<T>[], item: OdooModel, remote: any) => {
//   try {
//     console.debug('syncItemToRemote', model, stringify(item), stringify(remote));

//     const key = getKey(remote) || getKey(item);

//     const src: any = {}
//     for (const prop of props) {
//       const [srcProp, itemProp, toItemVal, toSrcVal] = prop;
//       src[srcProp] = (toSrcVal||toItemVal)((item as any)[itemProp]);
//     }

//     if (!remote) {
//       if (item.src) {
//         console.info('delete remote', stringify(item));
//         pbDeleted.push({ id: item.id });
//         await coll.delete(item.id);
//         return;
//       }
//       console.info('create remote', item.id, stringify(src));
//       odooCreated.push(src);
//       const response = await odooCreate(model, src);
//       console.info('create remote response', stringify(response));
//       // T0D0 src.id
//       // log('D', 'create remote response', stringify(response));
//       await coll.update(item.id, { src } as Partial<T>, { select: getSelect(props) });
//       return;
//     }

//     if (item.deleted) {
//       console.info('delete remote', item.id, key);
//       odooDeleted.push({ key });
//       await odooDelete(model, key);
//       await coll.delete(item.id);
//       return;
//     }

//     const remoteSrc = remote.src;
//     const srcUpdate: any = {};
//     for (const prop of props) {
//       const [srcProp] = prop;
//       if (!isEqual(src[srcProp], remoteSrc[srcProp])) {
//         srcUpdate[srcProp] = src[srcProp];
//       }
//     }
//     if (Object.keys(srcUpdate).length > 0) {
//       console.info('odooUpdate', getKey(item), ...Object.keys(srcUpdate));
//       odooUpdated.push({ key: getKey(item), ...srcUpdate });
//       await odooUpdate(model, getKey(item), srcUpdate);
//       const update = { src: { ...remoteSrc, ...srcUpdate } } as Partial<T>;
//       await coll.update(item.id, update);
//     }
//   }
//   catch (error) {
//     console.debug('error item', item.id, stringify(remote), stringify(item), toErr(error).message);
//   }
// }

// await odooSync<CategoryModel>("product.category", categoryColl, categoryProps);

// setProgress(30);
// setResult({ pbCreated, pbUpdated, pbDeleted, odooCreated, odooUpdated, odooDeleted });

// await odooSync<ProductModel>("product.product", productColl, productProps);

// setProgress(40);
// setResult({ pbCreated, pbUpdated, pbDeleted, odooCreated, odooUpdated, odooDeleted });









// const oList = <T, U>(
//   model: string,
//   fields: (keyof U)[],
//   map: (item: U) => T,
// ) =>
//   oCall(
//     model,
//     "search_read",
//     [],
//     {
//       fields,
//       limit: 99999,
//     },
//   ).then((r) => (toArray(r.result) as U[]).map(map));

// interface RefCategory extends ModelUpdate<CategoryModel> { ref: string; }
// interface RefProduct extends ModelUpdate<ProductModel> { ref: string; }

// const boolProp: PropFun = [b => isBool(b) ? (b ? 1 : 0) : undefined, toBool];
// const nbrProp: PropFun = [toNbr, toNbr];
// const colorProp: PropFun = [toStr, toStr];
// const priceProp: PropFun = [toStr, toNbr];
// const anyProp: PropFun = [v => v, v => v];

// const createConverter = <T>(itemName: string, props: Prop<T>[]): [
//   (from: any) => T,
//   (from: Partial<T>) => any,
// ] => {
//   console.debug('create converter', itemName, props);
//   const convert = <From, To>(i: 0|1) => {
//     const j = i ? 0 : 1;
//     const map: Record<string, [string, (v: any) => any]> = Object.fromEntries(
//       props.map(p => [p[j] as string, [p[i] as string, p[2][i]]])
//     );
//     return (from: From): To => {
//       const to: any = {};
//       for (const p in from) {
//         const v = (from as any)[p];
//         try {
//           const prop = map[p];
//           if (!prop) console.warn('convert unknown prop', itemName, p);
//           else {
//             const v2 = prop[1](v);
//             if (v2 !== undefined) {
//               to[prop[0]] = v2;
//             }
//           }
//         }
//         catch (err) {
//           console.warn('convert error', itemName, p, v, err);
//         }
//       }
//       return to;
//     }
//   }
//   return [convert<any, T>(1), convert<Partial<T>, any>(0)];
// }

// const srcCategories = () => 
  
//   r => ({
//   name: toStr(r.name),
//   desc: toStr(r.complete_name),
//   enabled: true,
//   remote: { odoo: r },
//   group,
  
  
  
//   {
//   id: '',
//   name: 'name',
//   complete_name: 'desc',
//   parent_id: '',
// }, {
//   enabled: true,
//   group,
// });


//   const odooItemList = await odooList("product.category", [
//     "id",
//     "name",
//     "complete_name",
//     "parent_id",
//   ], (p): ModelUpdate<CategoryModel> => ({
//     name: toStr(p.name),
//     desc: toStr(p.complete_name),
//     enabled: true,
//     remote: { odoo: p },
//     group,
//   }));
//   // remotes.forEach(p => p.parent = remoteByRef[p.parent||'']?.id);
//   const odooItems = by(odooItemList, getOdooId);

//   const itemList = await categoryColl.find({ group, ref: ["!=", ""] });
//   const items = by(itemList, getOdooId);

//   for (const [odooId, remote] of Object.entries(odooItems)) {
//     const item = items[odooId];

//     if (!item) {
//       items[odooId] = await categoryColl.create(remote);
//       continue;
//     }

//     if (!isEqual(item.remote.odoo, remote.remote.odoo)) {
//       items[odooId] = await categoryColl.update(item.id, remote);
//       continue;
//     }
//   }

//   // for (const [odooId, item] of Object.entries(items)) {
//   //   const odooItem = odooItems[odooId];

//   //   if (!odooItem) {
//   //     await categoryColl.create(remote);
//   //     continue;
//   //   }

//   //   if (!isEqual(item.remote.odoo, remote.remote.odoo)) {
//   //     await categoryColl.update(item.id, remote);
//   //     continue;
//   //   }
//   // }

//   // const itemWithoutRef = items.filter(p => !p.ref);
//   // const refs = Object.keys(deleteKey({ ...remoteByRef, ...itemByRef }, ''));
//   // const byRef = by(refs, null, ref => [remoteByRef[ref], itemByRef[ref]]);
//   // return byRef;
// };


// // const srcProducts = async () => {
// //     const res = await odooCall(
// //         "product.product",
// //         "search_read",
// //         [],
// //         {
// //             fields: [
// //                 "id",
// //                 "name",
// //                 "description",
// //                 "list_price",
// //                 "product_tmpl_id",
// //                 "categ_id",
// //                 "nbr_reordering_rules",
// //             ], // , "image_1920" image_256
// //             limit: 99999
// //         }
// //     );

// //     const remotes = toArray(res.result).map((p: any): OdooProduct => ({
// //         ref: p.id,
// //         data: p,
// //         desc: p.description,
// //         name: p.name,
// //         enabled: p.active,
// //         price: p.list_price,
// //         group,
// //         order: p.nbr_reordering_rules,
// //         srchronized: apiNow(),
// //         // category: p.categ_id,
// //         // image: '',
// //     }));
// //     const remoteByRef = by(remotes, p => p.ref);

// //     const items = (await productColl.find({ group })).filter(p => p.ref) as OdooProduct[];
// //     const itemByRef = by(items, p => p.ref);

// //     const byRef = by({ ...remoteByRef, ...itemByRef }, p => p.ref, p => [remoteByRef[p.ref], itemByRef[p.ref]]);

// //     return byRef;
// // }

// // const itemByRef = await srcProducts();

// setProgress(60);

// // const odooProducts = by(toArray(odooProducts.result).map((p: any): ModelUpdate<ProductModel> => ({
// //     ref: p.id,
// //     data: p,
// //     desc: p.description,
// //     name: p.name,
// //     enabled: p.active,
// //     price: p.list_price,
// //     group,
// //     order: p.nbr_reordering_rules,
// //     srchronized: apiNow(),
// //     // category: p.categ_id,
// //     // image: '',
// // })), p => p.ref);

// // // add sale
// // const addSale = async () => {
// //     const sale = await odooCall(
// //         "sale.order",
// //         "create",
// //         [
// //             {
// //                 partner_id: 3,
// //                 order_line: [
// //                     [0, 0, {
// //                         product_id: 8,
// //                         product_uom_qty: 1,
// //                         price_unit: 9.90
// //                     }]
// //                 ]
// //             }
// //         ]
// //     )
// // }

// // log("D", "odooProducts", odooProducts);
// // const url = `https://${credential.db}.odoo.com/web/session/authenticate`
// // log("D", url, { jsonrpc: "2.0", params: credential });
// // const odooAuth = await fetch(url, {
// //     method: 'POST',
// //     body: JSON.stringify({ jsonrpc: "2.0", params: credential }),
// //     headers: {
// //         'Content-Type': 'application/json'
// //     }
// // }).then(res => res.json());

