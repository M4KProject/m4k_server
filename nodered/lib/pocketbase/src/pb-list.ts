import { NodeAPI, Node, NodeDef } from 'node-red';
import { pbAutoAuth, requiredError } from './common';

export interface PBListNodeDef extends NodeDef {
    name: string;
    collection: string;
    page: number;
    perPage: number;
    filter: string;
    sort: string;
    mode: 'page' | 'items' | 'split';
}

module.exports = (RED: NodeAPI) => {
    const PBListNode = function(this: Node, def: PBListNodeDef) {
        RED.nodes.createNode(this, def);

        this.on('input', async (msg: any) => {
            try {
                const pb = await pbAutoAuth(this, msg);
                
                const collection = def.collection || msg.collection;
                const page = def.page || msg.page || 1;
                const perPage = def.perPage || msg.perPage || 50;
                const filter = def.filter || msg.filter || '';
                const sort = def.sort || msg.sort || '';
                const mode = def.mode || msg.mode || 'page';

                if (!collection) throw requiredError('Collection');

                this.debug(`PB List: ${collection} page=${page} perPage=${perPage} filter='${filter}' sort='${sort}' mode=${mode}`);

                const result = await pb.collection(collection).getList(page, perPage, {
                    filter,
                    sort,
                    expand: msg.expand || ''
                });

                msg.pb = pb;
                if (mode === 'page') {
                    msg.payload = result;
                    this.send(msg);
                } else if (mode === 'items') {
                    msg.payload = result.items;
                    this.send(msg);
                } else if (mode === 'split') {
                    result.items.forEach((item) => {
                        const newMsg = RED.util.cloneMessage(msg);
                        newMsg.payload = item;
                        this.send(newMsg);
                    });
                }

            } catch (error) {
                this.error(`PB List failed: ${error}`, msg);
            }
        });
    }
    
    RED.nodes.registerType("pb-list", PBListNode);
};