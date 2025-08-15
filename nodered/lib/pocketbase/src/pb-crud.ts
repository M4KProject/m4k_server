import { NodeAPI, Node, NodeDef } from 'node-red';
import PocketBase from 'pocketbase';

interface PBCrudNodeConfig extends NodeDef {
    name: string;
    operation: string;
    collection: string;
}

module.exports = function(RED: NodeAPI) {
    function PBCrudNode(this: Node, config: PBCrudNodeConfig) {
        RED.nodes.createNode(this, config);
        const node = this;
        
        node.on('input', async function(msg: any) {
            try {
                const pb = msg.pb || new PocketBase(process.env.PB_API_URL);
                const collection = config.collection || msg.collection;
                const operation = config.operation || msg.operation || 'getList';
                
                if (!collection) {
                    node.error("Collection name required", msg);
                    return;
                }
                
                let result;
                
                switch (operation) {
                    case 'getList':
                        result = await pb.collection(collection).getList(1, 50);
                        break;
                    case 'getOne':
                        result = await pb.collection(collection).getOne(msg.id);
                        break;
                    case 'create':
                        result = await pb.collection(collection).create(msg.data);
                        break;
                    case 'update':
                        result = await pb.collection(collection).update(msg.id, msg.data);
                        break;
                    case 'delete':
                        result = await pb.collection(collection).delete(msg.id);
                        break;
                    default:
                        node.error(`Unknown operation: ${operation}`, msg);
                        return;
                }
                
                msg.payload = result;
                node.send(msg);
                
            } catch (error) {
                node.error(`CRUD operation failed: ${error}`, msg);
            }
        });
    }
    
    RED.nodes.registerType("pb-crud", PBCrudNode);
};