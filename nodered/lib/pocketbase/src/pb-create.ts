import { NodeAPI, Node, NodeDef } from 'node-red';
import { pbAutoAuth, requiredError } from './common';

export interface PBCreateNodeDef extends NodeDef {
    name: string;
    collection: string;
    expand: string;
    json: string;
}

module.exports = (RED: NodeAPI) => {
    const PBCreateNode = function(this: Node, def: PBCreateNodeDef) {
        RED.nodes.createNode(this, def);

        this.on('input', async (msg: any) => {
            try {
                const pb = await pbAutoAuth(this, msg);
                
                let data = msg.payload;
                if (def.json && def.json.trim()) {
                    try {
                        data = JSON.parse(def.json);
                    } catch (jsonError) {
                        throw new Error(`Invalid JSON in configuration: ${jsonError}`);
                    }
                }
                
                const collection = msg.collection || def.collection || '';
                const expand = msg.expand || def.expand || '';

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