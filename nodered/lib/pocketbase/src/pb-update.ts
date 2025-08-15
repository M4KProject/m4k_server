import { NodeAPI, Node, NodeDef } from 'node-red';
import { pbAutoAuth, requiredError } from './common.ts';

export interface PBUpdateNodeDef extends NodeDef {
    name: string;
    collection: string;
    recordId: string;
    expand: string;
}

module.exports = (RED: NodeAPI) => {
    const PBUpdateNode = function(this: Node, def: PBUpdateNodeDef) {
        RED.nodes.createNode(this, def);

        this.on('input', async (msg: any) => {
            try {
                const pb = await pbAutoAuth(this, msg);
                
                const collection = def.collection || msg.collection;
                const id = def.recordId || msg.id;
                const data = msg.data || msg.payload;
                const expand = def.expand || msg.expand || '';

                if (!collection) throw requiredError('Collection');
                if (!id) throw requiredError('Record ID');
                if (!data) throw requiredError('Record data');

                this.debug(`PB Update: ${collection}/${id} expand='${expand}'`);

                const result = await pb.collection(collection).update(id, data, { expand });

                msg.payload = result;
                msg.pb = pb;
                this.send(msg);

            } catch (error) {
                this.error(`PB Update failed: ${error}`, msg);
            }
        });
    }
    
    RED.nodes.registerType("pb-update", PBUpdateNode);
};