import { NodeAPI, Node, NodeDef } from 'node-red';
import { pbAutoAuth, requiredError } from './common.ts';

export interface PBCreateNodeDef extends NodeDef {
    name: string;
    collection: string;
    expand: string;
}

module.exports = (RED: NodeAPI) => {
    const PBCreateNode = function(this: Node, def: PBCreateNodeDef) {
        RED.nodes.createNode(this, def);

        this.on('input', async (msg: any) => {
            try {
                const pb = await pbAutoAuth(this, msg);
                
                const collection = def.collection || msg.collection;
                const data = msg.data || msg.payload;
                const expand = def.expand || msg.expand || '';

                if (!collection) throw requiredError('Collection');
                if (!data) throw requiredError('Record data');

                this.debug(`PB Create: ${collection} expand='${expand}'`);

                const result = await pb.collection(collection).create(data, { expand });

                msg.payload = result;
                msg.pb = pb;
                this.send(msg);

            } catch (error) {
                this.error(`PB Create failed: ${error}`, msg);
            }
        });
    }
    
    RED.nodes.registerType("pb-create", PBCreateNode);
};