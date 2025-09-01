import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatListModule } from '@angular/material/list';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatGridListModule } from '@angular/material/grid-list';
import { Router } from '@angular/router';
import { GanjeApiService } from '../../core/services/ganje-api.service';
import { AuthService } from '../../core/services/auth.service';

interface DashboardStats {
  totalRepositories: number;
  totalArtifacts: number;
  totalUsers: number;
  storageUsed: string;
  recentActivity: any[];
}

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [
    CommonModule,
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatListModule,
    MatProgressBarModule,
    MatProgressSpinnerModule,
    MatGridListModule
  ],
  template: `
    <div class="dashboard-container">
      <div class="dashboard-header">
        <h1>Dashboard</h1>
        <p>Welcome to Ganje Admin Portal</p>
      </div>

      <div class="stats-grid" *ngIf="!loading">
        <mat-card class="stat-card">
          <mat-card-content>
            <div class="stat-content">
              <mat-icon class="stat-icon repositories">archive</mat-icon>
              <div class="stat-info">
                <h3>{{ stats.totalRepositories }}</h3>
                <p>Repositories</p>
              </div>
            </div>
          </mat-card-content>
        </mat-card>

        <mat-card class="stat-card">
          <mat-card-content>
            <div class="stat-content">
              <mat-icon class="stat-icon artifacts">inventory</mat-icon>
              <div class="stat-info">
                <h3>{{ stats.totalArtifacts }}</h3>
                <p>Artifacts</p>
              </div>
            </div>
          </mat-card-content>
        </mat-card>

        <mat-card class="stat-card" *ngIf="isAdmin">
          <mat-card-content>
            <div class="stat-content">
              <mat-icon class="stat-icon users">people</mat-icon>
              <div class="stat-info">
                <h3>{{ stats.totalUsers }}</h3>
                <p>Users</p>
              </div>
            </div>
          </mat-card-content>
        </mat-card>

        <mat-card class="stat-card">
          <mat-card-content>
            <div class="stat-content">
              <mat-icon class="stat-icon storage">storage</mat-icon>
              <div class="stat-info">
                <h3>{{ stats.storageUsed }}</h3>
                <p>Storage Used</p>
              </div>
            </div>
          </mat-card-content>
        </mat-card>
      </div>

      <div class="dashboard-actions">
        <mat-card>
          <mat-card-header>
            <mat-card-title>Quick Actions</mat-card-title>
          </mat-card-header>
          <mat-card-content>
            <div class="action-buttons">
              <button mat-raised-button color="primary" (click)="navigateTo('/repositories')">
                <mat-icon>archive</mat-icon>
                Manage Repositories
              </button>
              <button mat-raised-button color="accent" (click)="navigateTo('/artifacts')">
                <mat-icon>inventory</mat-icon>
                Browse Artifacts
              </button>
              <button mat-raised-button color="warn" *ngIf="isAdmin" (click)="navigateTo('/dex-config')">
                <mat-icon>settings</mat-icon>
                Dex Configuration
              </button>
              <button mat-raised-button *ngIf="isAdmin" (click)="navigateTo('/users')">
                <mat-icon>people</mat-icon>
                User Management
              </button>
            </div>
          </mat-card-content>
        </mat-card>
      </div>

      <div class="recent-activity" *ngIf="stats.recentActivity && stats.recentActivity.length > 0">
        <mat-card>
          <mat-card-header>
            <mat-card-title>Recent Activity</mat-card-title>
          </mat-card-header>
          <mat-card-content>
            <mat-list>
              <mat-list-item *ngFor="let activity of stats.recentActivity">
                <mat-icon matListItemIcon>{{ getActivityIcon(activity.type) }}</mat-icon>
                <div matListItemTitle>{{ activity.description }}</div>
                <div matListItemLine>{{ activity.timestamp | date:'short' }}</div>
              </mat-list-item>
            </mat-list>
          </mat-card-content>
        </mat-card>
      </div>

      <div class="loading-container" *ngIf="loading">
        <mat-spinner diameter="50"></mat-spinner>
        <p>Loading dashboard data...</p>
      </div>

      <div *ngIf="errorMessage" class="error-message">
        <mat-icon color="warn">error</mat-icon>
        {{ errorMessage }}
      </div>
    </div>
  `,
  styles: [`
    .dashboard-container {
      padding: 2rem;
      max-width: 1200px;
      margin: 0 auto;
    }

    .dashboard-header {
      margin-bottom: 2rem;
    }

    .dashboard-header h1 {
      margin: 0 0 0.5rem;
      color: var(--text-basic-color);
    }

    .dashboard-header p {
      margin: 0;
      color: var(--text-hint-color);
    }

    .stats-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
      gap: 1.5rem;
      margin-bottom: 2rem;
    }

    .stat-card {
      transition: transform 0.2s ease;
    }

    .stat-card:hover {
      transform: translateY(-2px);
    }

    .stat-content {
      display: flex;
      align-items: center;
      gap: 1rem;
    }

    .stat-icon {
      font-size: 2.5rem;
      width: 60px;
      height: 60px;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
    }

    .stat-icon.repositories {
      background: rgba(61, 90, 254, 0.1);
      color: #3d5afe;
    }

    .stat-icon.artifacts {
      background: rgba(0, 200, 83, 0.1);
      color: #00c853;
    }

    .stat-icon.users {
      background: rgba(255, 193, 7, 0.1);
      color: #ffc107;
    }

    .stat-icon.storage {
      background: rgba(156, 39, 176, 0.1);
      color: #9c27b0;
    }

    .stat-info h3 {
      margin: 0 0 0.25rem;
      font-size: 2rem;
      font-weight: 600;
    }

    .stat-info p {
      margin: 0;
      color: var(--text-hint-color);
      font-size: 0.875rem;
    }

    .dashboard-actions {
      margin-bottom: 2rem;
    }

    .action-buttons {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
      gap: 1rem;
    }

    .action-buttons button {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      justify-content: center;
      padding: 1rem;
    }

    .recent-activity {
      margin-bottom: 2rem;
    }

    .activity-item {
      display: flex;
      align-items: center;
      gap: 1rem;
      padding: 0.5rem 0;
    }

    .activity-content p {
      margin: 0 0 0.25rem;
      font-weight: 500;
    }

    .activity-content small {
      color: var(--text-hint-color);
    }

    .loading-container {
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      padding: 4rem 2rem;
      text-align: center;
    }

    .loading-container p {
      margin-top: 1rem;
      color: var(--text-hint-color);
    }

    @media (max-width: 768px) {
      .dashboard-container {
        padding: 1rem;
      }

      .stats-grid {
        grid-template-columns: 1fr;
      }

      .action-buttons {
        grid-template-columns: 1fr;
      }
    }
  `]
})
export class DashboardComponent implements OnInit {
  loading = true;
  errorMessage = '';
  isAdmin = false;
  
  stats: DashboardStats = {
    totalRepositories: 0,
    totalArtifacts: 0,
    totalUsers: 0,
    storageUsed: '0 GB',
    recentActivity: []
  };

  constructor(
    private ganjeApiService: GanjeApiService,
    private authService: AuthService,
    private router: Router
  ) {}

  async ngOnInit() {
    this.isAdmin = this.authService.hasPermission('admin');
    await this.loadDashboardData();
  }

  async loadDashboardData() {
    try {
      this.loading = true;
      this.errorMessage = '';

      // Load repositories
      const repositories = await this.ganjeApiService.getRepositories().toPromise();
      this.stats.totalRepositories = repositories?.length || 0;

      // Calculate total artifacts and storage from repositories
      let totalArtifacts = 0;
      let totalSize = 0;
      
      if (repositories) {
        repositories.forEach(repo => {
          totalArtifacts += repo.total_artifacts || 0;
          totalSize += repo.total_size || 0;
        });
      }

      this.stats.totalArtifacts = totalArtifacts;
      this.stats.storageUsed = this.formatBytes(totalSize);

      // Mock recent activity for now
      this.stats.recentActivity = [
        {
          type: 'upload',
          description: 'New artifact uploaded to maven-central',
          timestamp: new Date(Date.now() - 1000 * 60 * 30) // 30 minutes ago
        },
        {
          type: 'repository',
          description: 'Repository npm-registry created',
          timestamp: new Date(Date.now() - 1000 * 60 * 60 * 2) // 2 hours ago
        },
        {
          type: 'user',
          description: 'New user registered',
          timestamp: new Date(Date.now() - 1000 * 60 * 60 * 4) // 4 hours ago
        }
      ];

      // Mock user count for admin users
      if (this.isAdmin) {
        this.stats.totalUsers = 15; // This would come from a real API
      }

    } catch (error) {
      console.error('Failed to load dashboard data:', error);
      this.errorMessage = 'Failed to load dashboard data. Please try again.';
    } finally {
      this.loading = false;
    }
  }

  navigateTo(path: string) {
    this.router.navigate([path]);
  }

  getActivityIcon(type: string): string {
    switch (type) {
      case 'upload':
        return 'upload-outline';
      case 'repository':
        return 'archive-outline';
      case 'user':
        return 'person-add-outline';
      default:
        return 'activity-outline';
    }
  }

  private formatBytes(bytes: number): string {
    if (bytes === 0) return '0 GB';
    
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }
}
