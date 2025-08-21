import { CanActivateFn } from '@angular/router';
import { inject } from '@angular/core';
import { AuthService } from '../core/services/auth';
import { Router } from '@angular/router';

export const authGuardGuard: CanActivateFn = (route, state) => {
  const auth = inject(AuthService);
  const router = inject(Router);
  if (auth.isAuthenticated()) return true;
  router.navigate(['/auth/login'], { queryParams: { returnUrl: state.url } });
  return false;
};
