import { Routes } from '@angular/router';
import { authGuard } from '@core/auth/auth.guard';

export enum RoutesPath {
  Home = '',
  Dashboard = 'dashboard',
  SignIn = 'sign-in',
  SignUp = 'sign-up',
  Profile = 'profile',
}

export function PathWithSlash(route: RoutesPath): string {
  return '/' + route;
}

export interface MenuItem {
  path: RoutesPath;
  title: string;
  icon?: string;
}

export const MENU_ITEMS: MenuItem[] = [
  { path: RoutesPath.Home, title: 'О проекте', icon: 'home' },
  { path: RoutesPath.Dashboard, title: 'Дашбоард', icon: 'dashboard' },
];

export const routes: Routes = [
  {
    path: RoutesPath.Home,
    loadComponent: () => import('./pages/home/home.component').then(c => c.HomeComponent),
    pathMatch: 'full',
    canActivate: [authGuard],
  },
  {
    path: RoutesPath.Dashboard,
    loadComponent: () => import('./pages/dashboard/dashboard.component').then(c => c.DashboardComponent),
    canActivate: [authGuard],
  },
  {
    path: RoutesPath.SignIn,
    loadComponent: () => import('./pages/auth/sign-in/sign-in.component').then(c => c.SignInComponent),
  },
  {
    path: RoutesPath.SignUp,
    loadComponent: () => import('./pages/auth/sign-up/sign-up.component').then(c => c.SignUpComponent),
  },
  { path: '**', redirectTo: RoutesPath.Dashboard },
];
