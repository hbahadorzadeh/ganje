import { Injectable, Inject, PLATFORM_ID } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BehaviorSubject, Observable } from 'rxjs';
import { JwtHelperService } from '@auth0/angular-jwt';
import { Router } from '@angular/router';
import { isPlatformBrowser } from '@angular/common';

export interface User {
  id: number;
  username: string;
  email: string;
  realms: string[];
  active: boolean;
}

export interface LoginResponse {
  token: string;
  user: User;
}

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private readonly TOKEN_KEY = 'ganje_token';
  private readonly USER_KEY = 'ganje_user';
  private currentUserSubject = new BehaviorSubject<User | null>(null);
  public currentUser$ = this.currentUserSubject.asObservable();

  constructor(
    private http: HttpClient,
    private jwtHelper: JwtHelperService,
    private router: Router,
    @Inject(PLATFORM_ID) private platformId: Object
  ) {
    this.loadStoredUser();
  }

  private loadStoredUser(): void {
    if (!isPlatformBrowser(this.platformId)) {
      return; // Skip localStorage access during SSR
    }
    
    const token = this.getToken();
    const userStr = localStorage.getItem(this.USER_KEY);
    
    if (token && !this.jwtHelper.isTokenExpired(token) && userStr) {
      const user = JSON.parse(userStr);
      this.currentUserSubject.next(user);
    }
  }

  login(username: string, password: string): Observable<LoginResponse> {
    return new Observable(observer => {
      // For Dex integration, redirect to Dex OAuth flow
      this.redirectToDex();
    });
  }

  redirectToDex(): void {
    if (!isPlatformBrowser(this.platformId)) {
      return; // Skip redirect during SSR
    }
    
    // Dex OIDC authorization endpoint
    const dexBaseUrl = 'http://localhost:5556'; // Will be configured via environment
    const dexUrl = `${dexBaseUrl}/auth`;
    const clientId = 'ganje-admin-portal';
    const redirectUri = encodeURIComponent(window.location.origin + '/auth/callback');
    const scope = 'openid profile email groups';
    const state = this.generateState();
    
    // Store state for validation
    if (isPlatformBrowser(this.platformId)) {
      localStorage.setItem('oauth_state', state);
    }
    
    const authUrl = `${dexUrl}?client_id=${clientId}&redirect_uri=${redirectUri}&response_type=code&scope=${scope}&state=${state}`;
    window.location.href = authUrl;
  }

  private generateState(): string {
    return Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15);
  }

  handleAuthCallback(code: string): Observable<LoginResponse> {
    // Use backend URL directly since proxy isn't working with new Angular CLI
    const backendUrl = 'http://localhost:8080';
    return this.http.post<LoginResponse>(`${backendUrl}/api/v1/auth/callback`, { code });
  }

  setToken(token: string, user: User): void {
    if (isPlatformBrowser(this.platformId)) {
      localStorage.setItem(this.TOKEN_KEY, token);
      localStorage.setItem(this.USER_KEY, JSON.stringify(user));
    }
    this.currentUserSubject.next(user);
  }

  getToken(): string | null {
    if (!isPlatformBrowser(this.platformId)) {
      return null;
    }
    return localStorage.getItem(this.TOKEN_KEY);
  }

  getCurrentUser(): User | null {
    return this.currentUserSubject.value;
  }

  hasPermission(permission: string): boolean {
    const user = this.getCurrentUser();
    if (!user || !user.realms) return false;
    
    // Check if user has admin realm
    if (user.realms.includes('admins')) return true;
    
    // Check specific permissions based on realms
    switch (permission) {
      case 'read':
        return user.realms.some(realm => ['developers', 'readonly', 'admins'].includes(realm));
      case 'write':
        return user.realms.some(realm => ['developers', 'admins'].includes(realm));
      case 'admin':
        return user.realms.includes('admins');
      default:
        return false;
    }
  }

  isAuthenticated(): boolean {
    const token = this.getToken();
    return token != null && !this.jwtHelper.isTokenExpired(token);
  }


  logout(): void {
    if (isPlatformBrowser(this.platformId)) {
      localStorage.removeItem(this.TOKEN_KEY);
      localStorage.removeItem(this.USER_KEY);
    }
    this.currentUserSubject.next(null);
    this.router.navigate(['/login']);
  }
}
