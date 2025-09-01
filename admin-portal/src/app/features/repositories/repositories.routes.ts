import { Routes } from '@angular/router';

export const repositoriesRoutes: Routes = [
  {
    path: '',
    loadComponent: () => import('./repository-list/repository-list.component').then(m => m.RepositoryListComponent)
  },
  {
    path: 'create',
    loadComponent: () => import('./repository-create/repository-create.component').then(m => m.RepositoryCreateComponent)
  },
  {
    path: ':name',
    loadComponent: () => import('./repository-detail/repository-detail.component').then(m => m.RepositoryDetailComponent)
  },
  {
    path: ':name/edit',
    loadComponent: () => import('./repository-edit/repository-edit.component').then(m => m.RepositoryEditComponent)
  }
];
