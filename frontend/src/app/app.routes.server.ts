import { RenderMode, ServerRoute } from '@angular/ssr';
import { RoutesPath } from './app.routes';

export const serverRoutes: ServerRoute[] = [
  { path: RoutesPath.Home, renderMode: RenderMode.Prerender },
  { path: RoutesPath.Dashboard, renderMode: RenderMode.Prerender },
  { path: '**', renderMode: RenderMode.Prerender },
];
