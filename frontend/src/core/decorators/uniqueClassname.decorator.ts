import { makeid } from "../tools/random.tools";

var classnameDictionnay:Record<string, boolean> = {};

export function GetUniqueClassName(target: any): string {
    if (!target.hasOwnProperty('uniqueClassname')) {
        let curName = target.name;
        if (curName in classnameDictionnay) {
            curName = makeid(5) + Date.now();
        }
        target.uniqueClassname = curName;
        classnameDictionnay[curName] = true;
    }
    return target.uniqueClassname;
}
