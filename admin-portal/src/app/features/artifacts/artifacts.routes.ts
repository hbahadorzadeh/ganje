import { Routes } from '@angular/router';

export const artifactsRoutes: Routes = [
  {
    path: '',
    loadComponent: () => import('./artifact-list/artifact-list.component').then(m => m.ArtifactListComponent)
  }
];
