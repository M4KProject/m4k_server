import { NodeAPI, Node, NodeDef } from 'node-red';
import PocketBase from 'pocketbase';

interface PBAuthNodeConfig extends NodeDef {
    name: string;
    apiUrl: string;
    username: string;
    password: string;
}

module.exports = function(RED: NodeAPI) {
    function PBAuthNode(this: Node, config: PBAuthNodeConfig) {
        RED.nodes.createNode(this, config);
        const node = this;
        
        node.on('input', async function(msg: any) {
            try {
                const apiUrl = config.apiUrl || process.env.PB_API_URL;
                const username = config.username || process.env.PB_ADMIN_USERNAME;
                const password = config.password || process.env.PB_ADMIN_PASSWORD;
                
                if (!apiUrl || !username || !password) {
                    node.error("Missing API URL, username or password", msg);
                    return;
                }
                
                const pb = new PocketBase(apiUrl);
                const authData = await pb.collection('_superusers').authWithPassword(username, password);
                
                msg.payload = authData;
                msg.pb = pb;
                
                node.send(msg);
                
            } catch (error) {
                node.error(`Auth failed: ${error}`, msg);
            }
        });
    }
    
    RED.nodes.registerType("pb-auth", PBAuthNode);
};