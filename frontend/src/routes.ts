import { Route } from "./core/interfaces/routeconf.interface";
import { VerifyPage } from "./pages/verify/verify.page";
import { HomePage } from "./pages/home/home.page";

export const AppRoutes:Route[] = [
    {
        path: "/",
        pageComponent: HomePage
    },
    {
        path: "/verify",
        pageComponent: VerifyPage
    }
]