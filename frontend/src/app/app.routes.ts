import { Routes } from '@angular/router';

export enum RoutesPath {
  Home = '',
  Dashboard = 'dashboard'
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
  { path: RoutesPath.Home, loadComponent: () => import('./pages/home/home.component').then(c => c.HomeComponent), pathMatch: 'full' },
  { path: RoutesPath.Dashboard, loadComponent: () => import('./pages/dashboard/dashboard.component').then(c => c.DashboardComponent)},
  { path: '**', redirectTo: '' },
];
