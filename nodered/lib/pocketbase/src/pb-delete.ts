import { NodeAPI, Node, NodeDef } from 'node-red';
import { isString, pbAutoAuth, propError } from './common';

export interface PBDeleteNodeDef extends NodeDef {
    name: string;
    collection: string;
    recordId: string;
    confirm: boolean;
}

module.exports = (RED: NodeAPI) => {
    const PBDeleteNode = function(this: Node, def: PBDeleteNodeDef) {
        RED.nodes.createNode(this, def);

        this.on('input', async (msg: any) => {
            try {
                const pb = await pbAutoAuth(this, msg);
                
                const p = msg.payload || {};
                const collection = def.collection || msg.collection || p.collectionName;
                const id = def.recordId || msg.recordId || p.id;

                if (!isString(collection)) throw propError('Collection');
                if (!id) throw propError('Record ID');

                this.debug(`PB Delete: ${collection}/${id}`);

                const isDeleted = await pb.collection(collection).delete(id);
                msg.payload = isDeleted;
                msg.pb = pb;
                this.send(msg);

            } catch (error) {
                this.error(`PB Delete failed: ${error}`, msg);
            }
        });
    }
    
    RED.nodes.registerType("pb-delete", PBDeleteNode);
};