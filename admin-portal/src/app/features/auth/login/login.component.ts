import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { AuthService } from '../../../core/services/auth.service';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [
    CommonModule,
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatProgressSpinnerModule
  ],
  template: `
    <div class="login-container">
      <mat-card class="login-card">
        <mat-card-header>
          <mat-card-title>Ganje Admin Portal</mat-card-title>
          <mat-card-subtitle>Sign in to access the admin dashboard</mat-card-subtitle>
        </mat-card-header>
        
        <mat-card-content>
          <div class="login-methods">
            <button mat-raised-button color="primary" class="login-button"
                    (click)="loginWithDex()" 
                    [disabled]="loading">
              <mat-icon>lock</mat-icon>
              Sign in with Dex
              <mat-spinner *ngIf="loading" diameter="20" class="login-spinner"></mat-spinner>
            </button>
          </div>
          
          <div *ngIf="errorMessage" class="error-message">
            <mat-icon color="warn">error</mat-icon>
            {{ errorMessage }}
          </div>
        </mat-card-content>
      </mat-card>
    </div>
  `,
  styles: [`
    .login-container {
      display: flex;
      justify-content: center;
      align-items: center;
      min-height: 100vh;
      background: linear-gradient(135deg, #3f51b5 0%, #9c27b0 100%);
      padding: 2rem;
    }
    
    .login-card {
      width: 100%;
      max-width: 400px;
      box-shadow: 0 8px 16px rgba(0,0,0,0.3);
    }
    
    mat-card-header {
      text-align: center;
      padding: 2rem 2rem 1rem;
    }
    
    mat-card-title {
      font-size: 1.5rem;
      margin-bottom: 0.5rem;
    }
    
    .login-methods {
      margin: 2rem 0;
    }
    
    .login-button {
      width: 100%;
      height: 48px;
      font-size: 16px;
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 8px;
    }
    
    .login-spinner {
      margin-left: 8px;
    }
    
    .error-message {
      display: flex;
      align-items: center;
      gap: 8px;
      color: #f44336;
      background-color: #ffebee;
      padding: 12px;
      border-radius: 4px;
      margin-top: 1rem;
    }
  `]
})
export class LoginComponent {
  loading = false;
  errorMessage = '';

  constructor(
    private authService: AuthService,
    private router: Router
  ) {}

  async loginWithDex() {
    try {
      this.loading = true;
      this.errorMessage = '';
      
      // Redirect to Dex OAuth flow
      this.authService.redirectToDex();
    } catch (error) {
      this.errorMessage = 'Failed to initiate login. Please try again.';
      this.loading = false;
    }
  }
}
