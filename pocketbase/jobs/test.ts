const [_id, input] = Deno.args

const sleep = (ms: number) => new Promise(resolve => setTimeout(resolve, ms))

await sleep(5000);

console.log('progress', 10);

console.error('mon log erreur', 1);
console.error('mon log erreur', 2);

await sleep(5000);

console.log('progress', 30);

console.warn('mon log warn', 1);
console.warn('mon log warn', 2);

await sleep(5000);

console.log('mon log', 1);
console.log('mon log', 2);

console.log('progress', 50);

await sleep(5000);

// await sleep(1000);

if (input === 'error') {
    throw new Error("mon erreur")
}

if (input === 'timeout') {
    await sleep(10000);
}

console.log("result", input);