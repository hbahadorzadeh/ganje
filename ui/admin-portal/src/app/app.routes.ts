import { Routes } from '@angular/router';
import { Login } from './auth/login/login';
import { Callback } from './auth/callback/callback';
import { List as ReposListComponent } from './repositories/list/list';
import { Detail as RepoDetailComponent } from './repositories/detail/detail';
import { Search as SearchComponent } from './search/search/search';
import { authGuardGuard } from './auth/auth-guard-guard';

export const routes: Routes = [
  { path: 'auth/login', component: Login },
  { path: 'auth/callback', component: Callback },
  { path: 'repositories', component: ReposListComponent, canActivate: [authGuardGuard] },
  { path: 'repositories/:name', component: RepoDetailComponent, canActivate: [authGuardGuard] },
  { path: 'search', component: SearchComponent, canActivate: [authGuardGuard] },
  { path: '', pathMatch: 'full', redirectTo: 'repositories' },
  { path: '**', redirectTo: 'repositories' }
];
