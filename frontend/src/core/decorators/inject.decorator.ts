import { AppGlobals } from "../common/global.app";
import { GetUniqueClassName } from "./uniqueClassname.decorator";

export function inject(serviceClass: any) {
    let global = AppGlobals.getInstance();
    let serviceName = GetUniqueClassName(serviceClass)
    return function (constructor: any, paramName: string | symbol | undefined, paramPosition: number) {
        let constructorUniqueName = GetUniqueClassName(constructor)
        if (!Reflect.has(global.pageServiceMapping, constructorUniqueName)) {
            global.pageServiceMapping[constructorUniqueName] = [];
        }
        global.pageServiceMapping[constructorUniqueName][paramPosition] = serviceName;
    }
}