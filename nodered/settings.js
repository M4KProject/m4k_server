// deno-lint-ignore-file no-process-global

const {
    PORT,
    TZ,
    NODERED_USERNAME,
    NODERED_PASSWORD_HASH,
    PB_ADMIN_EMAIL,
    PB_ADMIN_PASSWORD,
    S3_BUCKET,
    S3_REGION,
    S3_ENDPOINT,
    S3_ACCESS_KEY,
    S3_SECRET_KEY,
} = process.env;

if (!NODERED_USERNAME) {
    console.warn("[Node-RED] WARNING: No admin username configured!");
}

if (!NODERED_PASSWORD_HASH) {
    console.warn("[Node-RED] WARNING: No admin password configured!");
}

module.exports = {
    uiPort: PORT || 1880,
    uiHost: "0.0.0.0",

    // Authentification Admin
    adminAuth: {
        type: "credentials",
        users: [{
            username: NODERED_USERNAME,
            password: NODERED_PASSWORD_HASH,
            permissions: "*"
        }]
    },

    timezone: TZ || "Europe/Paris",
    flowFile: 'flows.json',

    logging: {
        console: {
            level: "info",
            metrics: false,
            audit: false
        }
    },
    
    functionGlobalContext: {
        PB_ADMIN_EMAIL,
        PB_ADMIN_PASSWORD,
        S3_BUCKET,
        S3_REGION,
        S3_ENDPOINT,
        S3_ACCESS_KEY,
        S3_SECRET_KEY,
    },
};
