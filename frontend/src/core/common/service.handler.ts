import { GetUniqueClassName } from "../decorators/uniqueClassname.decorator";
import { AppGlobals } from "./global.app";

export class ServiceHandler {
    createInstance(service: any) {
        let global = AppGlobals.getInstance();
        let args: any = [];
        let serviceUniqueName = GetUniqueClassName(service);
        if (Reflect.has(global.pageServiceMapping, serviceUniqueName)) {
            global.pageServiceMapping[serviceUniqueName].forEach(serviceName => {
                args.push(global.services[serviceName]);
            });
        }
        return Reflect.construct(service, args);
    }
}