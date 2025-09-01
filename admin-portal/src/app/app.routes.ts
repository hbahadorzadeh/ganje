import { Routes } from '@angular/router';
import { authGuard, adminGuard } from './core/guards/auth.guard';

export const routes: Routes = [
  {
    path: '',
    redirectTo: '/dashboard',
    pathMatch: 'full'
  },
  {
    path: 'login',
    loadComponent: () => import('./features/auth/login/login.component').then(m => m.LoginComponent)
  },
  {
    path: 'auth/callback',
    loadComponent: () => import('./features/auth/callback/callback.component').then(m => m.CallbackComponent)
  },
  {
    path: 'dashboard',
    loadComponent: () => import('./features/dashboard/dashboard.component').then(m => m.DashboardComponent),
    canActivate: [authGuard]
  },
  {
    path: 'repositories',
    loadChildren: () => import('./features/repositories/repositories.routes').then(m => m.repositoriesRoutes),
    canActivate: [authGuard]
  },
  {
    path: 'artifacts',
    loadChildren: () => import('./features/artifacts/artifacts.routes').then(m => m.artifactsRoutes),
    canActivate: [authGuard]
  },
  {
    path: 'dex-config',
    loadChildren: () => import('./features/dex-config/dex-config.routes').then(m => m.dexConfigRoutes),
    canActivate: [adminGuard]
  },
  {
    path: 'users',
    loadChildren: () => import('./features/users/users.routes').then(m => m.usersRoutes),
    canActivate: [adminGuard]
  },
  {
    path: '**',
    redirectTo: '/dashboard'
  }
];
