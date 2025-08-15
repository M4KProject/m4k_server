import { NodeAPI, Node, NodeDef } from 'node-red';
import { pbAutoAuth, requiredError } from './common.ts';

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
                
                const collection = def.collection || msg.collection;
                const id = def.recordId || msg.id;
                const confirm = def.confirm || msg.confirm || false;

                if (!collection) throw requiredError('Collection');
                if (!id) throw requiredError('Record ID');

                if (confirm && !msg.confirmed) {
                    this.warn(`Delete operation requires confirmation. Set msg.confirmed = true to proceed.`);
                    
                    msg.payload = {
                        action: 'delete_confirmation_required',
                        collection,
                        id,
                        message: `Delete record ${id} from ${collection}?`
                    };
                    this.send(msg);
                    return;
                }

                this.debug(`PB Delete: ${collection}/${id}`);

                const result = await pb.collection(collection).delete(id);

                msg.payload = {
                    action: 'deleted',
                    collection,
                    id,
                    success: result,
                    timestamp: new Date().toISOString()
                };
                msg.pb = pb;
                this.send(msg);

            } catch (error) {
                this.error(`PB Delete failed: ${error}`, msg);
            }
        });
    }
    
    RED.nodes.registerType("pb-delete", PBDeleteNode);
};