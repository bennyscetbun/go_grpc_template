import { AppGlobals } from "../common/global.app";
import { PageConf } from "../interfaces/pageconf.interface";
import { GetUniqueClassName } from "./uniqueClassname.decorator";

export function PageComponent(config:PageConf){
    let global = AppGlobals.getInstance();
    return function(target: any){
        let componentName = GetUniqueClassName(target);
        global.templates[componentName] = config.template;
    }
}