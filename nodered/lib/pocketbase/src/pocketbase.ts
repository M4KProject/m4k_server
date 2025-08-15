import { NodeAPI, Node, NodeDef } from 'node-red';
import PocketBase from 'pocketbase';

interface PocketBaseNodeConfig extends NodeDef {
    apiUrl: string;
    username: string;
    password: string;
}

interface PocketBaseNode extends Node {
    pb?: PocketBase;
}

module.exports = function(RED: NodeAPI) {
    function PocketBaseNode(this: PocketBaseNode, config: PocketBaseNodeConfig) {
        RED.nodes.createNode(this, config);
        
        const node = this;
        
        // Initialize PocketBase client
        const apiUrl = config.apiUrl || process.env.PB_API_URL;
        if (!apiUrl) {
            node.error("API URL is required");
            return;
        }
        
        node.pb = new PocketBase(apiUrl);
        
        node.on('input', async function(msg: any) {
            try {
                const username = config.username || process.env.PB_ADMIN_USERNAME;
                const password = config.password || process.env.PB_ADMIN_PASSWORD;
                
                if (!username || !password) {
                    node.error("Username and password are required", msg);
                    return;
                }
                
                // Authenticate as superuser
                const authData = await node.pb!.collection('_superusers').authWithPassword(username, password);
                
                // Set authenticated client and auth data on message
                msg.pb = node.pb;
                msg.payload = authData;
                
                // Send to success output (port 1)
                node.send([msg, null]);
                
            } catch (error) {
                node.error(`Authentication failed: ${error}`, msg);
                
                // Send to error output (port 2)
                const errorMsg = { ...msg, error: error, payload: null };
                node.send([null, errorMsg]);
            }
        });
        
        node.on('close', function() {
            if (node.pb) {
                node.pb.authStore.clear();
            }
        });
    }
    
    RED.nodes.registerType("pocketbase", PocketBaseNode);
};