import { Routes } from '@angular/router';

export const dexConfigRoutes: Routes = [
  {
    path: '',
    loadComponent: () => import('./dex-config-list/dex-config-list.component').then(m => m.DexConfigListComponent)
  }
];
