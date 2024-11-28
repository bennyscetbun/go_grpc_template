import {MainConfig} from 'src/core/decorators/main.decorator';
import { AppRoutes } from 'src/routes';
import { ApiService } from 'src/services/api.service';

window.addEventListener('load', function() {
    @MainConfig({
        rootElement:'app',
        routes: AppRoutes,
        services:[ApiService]
    })
    class Main{};
})