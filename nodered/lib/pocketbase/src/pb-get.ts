import { NodeAPI, Node, NodeDef } from 'node-red';
import { pbAutoAuth, requiredError } from './common';

export interface PBGetNodeDef extends NodeDef {
    name: string;
    collection: string;
    recordId: string;
    expand: string;
}

module.exports = (RED: NodeAPI) => {
    const PBGetNode = function(this: Node, def: PBGetNodeDef) {
        RED.nodes.createNode(this, def);

        this.on('input', async (msg: any) => {
            try {
                const pb = await pbAutoAuth(this, msg);

                const p = msg.payload || {};
                const collection = msg.collection || def.collection || p.collectionName;
                const id = msg.recordId || def.recordId || p.id;
                const expand = msg.expand || def.expand || '';

                if (!collection) throw requiredError('Collection');
                if (!id) throw requiredError('Record ID');

                this.debug(`PB Get: ${collection}/${id} expand='${expand}'`);

                const result = await pb.collection(collection).getOne(id, { expand });

                msg.payload = result;
                msg.pb = pb;
                this.send(msg);

            } catch (error) {
                this.error(`PB Get failed: ${error}`, msg);
            }
        });
    }
    
    RED.nodes.registerType("pb-get", PBGetNode);
};