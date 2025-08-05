import { JobModel } from "../../common/api/index.ts";
import { toErr } from "../../common/helpers/err.ts";

export const job: JobModel = JSON.parse(Deno.args[0]);
export const pbAuth = JSON.parse(Deno.args[0]);

const stringify = (arg: any) => {
    try {
        return (
            (typeof arg === 'string') ? arg :
            (arg instanceof Error) ? String(toErr(arg)) :
            JSON.stringify(arg)
        );
    }
    catch (err) {
        return toErr(err).message;
    }
};

const _log = console.log;

const send = (...args: unknown[]) => _log(args.map(stringify).join('\t'));

export const setProgress = (progress: number) => send("progress", progress);

export const setResult = (result: any) => send("result", result);

export const initConsole = () => {
    Object.assign(console, {
        debug: (...args: any[]) => send('D', ...args),
        log: (...args: any[]) => send('I', ...args),
        info: (...args: any[]) => send('I', ...args),
        warn: (...args: any[]) => send('W', ...args),
        error: (...args: any[]) => send('E', ...args),
    });
}
