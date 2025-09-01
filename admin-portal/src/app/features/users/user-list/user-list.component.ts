import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatTableModule } from '@angular/material/table';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSnackBar } from '@angular/material/snack-bar';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatChipsModule } from '@angular/material/chips';
import { MatSlideToggleModule } from '@angular/material/slide-toggle';
import { MatDialog } from '@angular/material/dialog';
import { GanjeApiService } from '../../../core/services/ganje-api.service';
import { ErrorHandlerService } from '../../../core/services/error-handler.service';
import { LoadingService } from '../../../core/services/loading.service';
import { LoadingSpinnerComponent } from '../../../shared/components/loading-spinner/loading-spinner.component';
import { ErrorDisplayComponent } from '../../../shared/components/error-display/error-display.component';

export interface User {
  id: number;
  username: string;
  email: string;
  realms: string[];
  active: boolean;
  created_at: string;
  updated_at: string;
  last_login?: string;
}

@Component({
  selector: 'app-user-list',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatInputModule,
    MatFormFieldModule,
    MatTableModule,
    MatProgressSpinnerModule,
    LoadingSpinnerComponent,
    ErrorDisplayComponent,
    MatTooltipModule,
    MatChipsModule,
    MatSlideToggleModule
  ],
  template: `
    <div class="user-container">
      <div class="page-header">
        <div class="header-content">
          <h1>User Management</h1>
          <p>Manage users and their permissions</p>
        </div>
        <button mat-raised-button color="primary" (click)="createUser()">
          <mat-icon>person_add</mat-icon>
          Add User
        </button>
      </div>

      <mat-card>
        <mat-card-header>
          <div class="card-header-content">
            <mat-card-title>Users ({{ users.length }})</mat-card-title>
            <div class="search-container">
              <mat-form-field appearance="outline">
                <mat-label>Search users</mat-label>
                <input matInput [(ngModel)]="searchTerm" (input)="filterUsers()" placeholder="Search by name, email...">
                <mat-icon matSuffix>search</mat-icon>
              </mat-form-field>
            </div>
          </div>
        </mat-card-header>
        
        <mat-card-content>
          <div *ngIf="loading" class="loading-container">
            <mat-spinner></mat-spinner>
            <p>Loading users...</p>
          </div>

          <table mat-table [dataSource]="filteredUsers" *ngIf="!loading && filteredUsers.length > 0">
            <ng-container matColumnDef="username">
              <th mat-header-cell *matHeaderCellDef>Username</th>
              <td mat-cell *matCellDef="let user">
                <div class="user-info">
                  <div class="user-details">
                    <strong>{{ user.username }}</strong>
                    <div class="user-email">{{ user.email }}</div>
                  </div>
                  <mat-icon *ngIf="!user.active" color="warn" matTooltip="Inactive user">block</mat-icon>
                </div>
              </td>
            </ng-container>

            <ng-container matColumnDef="realms">
              <th mat-header-cell *matHeaderCellDef>Roles & Permissions</th>
              <td mat-cell *matCellDef="let user">
                <mat-chip-set>
                  <mat-chip *ngFor="let realm of user.realms" [color]="getRealmColor(realm)">
                    {{ realm | titlecase }}
                  </mat-chip>
                </mat-chip-set>
                <span *ngIf="user.realms.length === 0" class="no-realms">No roles assigned</span>
              </td>
            </ng-container>

            <ng-container matColumnDef="status">
              <th mat-header-cell *matHeaderCellDef>Status</th>
              <td mat-cell *matCellDef="let user">
                <div class="status-info">
                  <mat-slide-toggle 
                    [checked]="user.active"
                    (change)="toggleUserStatus(user, $event.checked)"
                    [disabled]="updatingUser === user.id">
                    {{ user.active ? 'Active' : 'Inactive' }}
                  </mat-slide-toggle>
                  <div class="last-login" *ngIf="user.last_login">
                    Last login: {{ user.last_login | date:'short' }}
                  </div>
                </div>
              </td>
            </ng-container>

            <ng-container matColumnDef="created_at">
              <th mat-header-cell *matHeaderCellDef>Created</th>
              <td mat-cell *matCellDef="let user">{{ user.created_at | date:'short' }}</td>
            </ng-container>

            <ng-container matColumnDef="actions">
              <th mat-header-cell *matHeaderCellDef>Actions</th>
              <td mat-cell *matCellDef="let user">
                <div class="action-buttons">
                  <button mat-icon-button (click)="viewUser(user)" matTooltip="View Details">
                    <mat-icon>visibility</mat-icon>
                  </button>
                  <button mat-icon-button (click)="editUser(user)" matTooltip="Edit User">
                    <mat-icon>edit</mat-icon>
                  </button>
                  <button mat-icon-button (click)="managePermissions(user)" matTooltip="Manage Permissions">
                    <mat-icon>security</mat-icon>
                  </button>
                  <button mat-icon-button color="warn" (click)="deleteUser(user)" matTooltip="Delete User">
                    <mat-icon>delete</mat-icon>
                  </button>
                </div>
              </td>
            </ng-container>

            <tr mat-header-row *matHeaderRowDef="displayedColumns"></tr>
            <tr mat-row *matRowDef="let row; columns: displayedColumns;"></tr>
          </table>

          <div *ngIf="!loading && filteredUsers.length === 0" class="empty-state">
            <mat-icon>people</mat-icon>
            <h3>No users found</h3>
            <p *ngIf="searchTerm">Try adjusting your search criteria</p>
            <p *ngIf="!searchTerm">Create your first user to get started</p>
          </div>

          <div *ngIf="errorMessage" class="error-message">
            <mat-icon color="warn">error</mat-icon>
            {{ errorMessage }}
          </div>
        </mat-card-content>
      </mat-card>

      <!-- User Statistics -->
      <div class="stats-grid">
        <mat-card class="stat-card">
          <mat-card-content>
            <div class="stat-content">
              <mat-icon class="stat-icon">people</mat-icon>
              <div class="stat-info">
                <div class="stat-value">{{ users.length }}</div>
                <div class="stat-label">Total Users</div>
              </div>
            </div>
          </mat-card-content>
        </mat-card>

        <mat-card class="stat-card">
          <mat-card-content>
            <div class="stat-content">
              <mat-icon class="stat-icon" color="primary">check_circle</mat-icon>
              <div class="stat-info">
                <div class="stat-value">{{ getActiveUsersCount() }}</div>
                <div class="stat-label">Active Users</div>
              </div>
            </div>
          </mat-card-content>
        </mat-card>

        <mat-card class="stat-card">
          <mat-card-content>
            <div class="stat-content">
              <mat-icon class="stat-icon" color="accent">admin_panel_settings</mat-icon>
              <div class="stat-info">
                <div class="stat-value">{{ getAdminUsersCount() }}</div>
                <div class="stat-label">Admin Users</div>
              </div>
            </div>
          </mat-card-content>
        </mat-card>

        <mat-card class="stat-card">
          <mat-card-content>
            <div class="stat-content">
              <mat-icon class="stat-icon" color="warn">block</mat-icon>
              <div class="stat-info">
                <div class="stat-value">{{ getInactiveUsersCount() }}</div>
                <div class="stat-label">Inactive Users</div>
              </div>
            </div>
          </mat-card-content>
        </mat-card>
      </div>
    </div>
  `,
  styles: [`
    .user-container {
      padding: 2rem;
      max-width: 1200px;
      margin: 0 auto;
    }

    .page-header {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      margin-bottom: 2rem;
    }

    .header-content h1 {
      margin: 0 0 0.5rem;
    }

    .header-content p {
      margin: 0;
      color: #666;
    }

    .card-header-content {
      display: flex;
      justify-content: space-between;
      align-items: center;
      width: 100%;
    }

    .search-container {
      width: 300px;
    }

    .loading-container {
      display: flex;
      flex-direction: column;
      align-items: center;
      padding: 2rem;
      gap: 1rem;
    }

    .user-info {
      display: flex;
      align-items: center;
      gap: 1rem;
    }

    .user-details {
      display: flex;
      flex-direction: column;
    }

    .user-email {
      font-size: 0.875rem;
      color: #666;
    }

    .no-realms {
      font-style: italic;
      color: #999;
    }

    .status-info {
      display: flex;
      flex-direction: column;
      gap: 0.25rem;
    }

    .last-login {
      font-size: 0.75rem;
      color: #666;
    }

    .action-buttons {
      display: flex;
      gap: 0.25rem;
    }

    .empty-state {
      text-align: center;
      padding: 3rem;
      color: #666;
    }

    .empty-state mat-icon {
      font-size: 4rem;
      width: 4rem;
      height: 4rem;
      margin-bottom: 1rem;
    }

    .error-message {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      color: #d32f2f;
      padding: 1rem;
      background: #ffebee;
      border-radius: 4px;
      margin-top: 1rem;
    }

    .stats-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
      gap: 1rem;
      margin-top: 2rem;
    }

    .stat-card {
      background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
    }

    .stat-content {
      display: flex;
      align-items: center;
      gap: 1rem;
    }

    .stat-icon {
      font-size: 2.5rem;
      width: 2.5rem;
      height: 2.5rem;
    }

    .stat-info {
      display: flex;
      flex-direction: column;
    }

    .stat-value {
      font-size: 2rem;
      font-weight: bold;
      line-height: 1;
    }

    .stat-label {
      font-size: 0.875rem;
      color: #666;
      margin-top: 0.25rem;
    }

    @media (max-width: 768px) {
      .page-header {
        flex-direction: column;
        gap: 1rem;
      }
      
      .card-header-content {
        flex-direction: column;
        gap: 1rem;
        align-items: flex-start;
      }
      
      .search-container {
        width: 100%;
      }

      .stats-grid {
        grid-template-columns: repeat(2, 1fr);
      }
    }
  `]
})
export class UserListComponent implements OnInit {
  users: User[] = [];
  filteredUsers: User[] = [];
  loading = true;
  errorMessage = '';
  searchTerm = '';
  updatingUser?: number;
  
  displayedColumns: string[] = ['username', 'realms', 'status', 'created_at', 'actions'];

  constructor(
    private router: Router,
    private snackBar: MatSnackBar,
    private dialog: MatDialog,
    private errorHandler: ErrorHandlerService,
    private loadingService: LoadingService,
    private ganjeApiService: GanjeApiService
  ) {}

  async ngOnInit() {
    await this.loadUsers();
  }

  private async loadUsers() {
    const loadingKey = 'users-list';
    this.loading = true;
    this.loadingService.setLoading(loadingKey, true);
    
    try {
      // Simulate API call with mock data
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      this.users = [
        {
          id: 1,
          username: 'admin',
          email: 'admin@ganje.io',
          realms: ['admin', 'developers'],
          active: true,
          created_at: '2024-01-15T10:30:00Z',
          updated_at: '2024-01-20T14:22:00Z',
          last_login: '2024-01-20T14:22:00Z'
        },
        {
          id: 2,
          username: 'developer1',
          email: 'dev1@company.com',
          realms: ['developers'],
          active: true,
          created_at: '2024-01-16T09:15:00Z',
          updated_at: '2024-01-19T16:45:00Z',
          last_login: '2024-01-19T16:45:00Z'
        },
        {
          id: 3,
          username: 'viewer1',
          email: 'viewer@company.com',
          realms: ['viewers'],
          active: false,
          created_at: '2024-01-17T11:20:00Z',
          updated_at: '2024-01-18T08:30:00Z',
          last_login: '2024-01-18T08:30:00Z'
        },
        {
          id: 4,
          username: 'tester1',
          email: 'tester@company.com',
          realms: ['testers', 'developers'],
          active: true,
          created_at: '2024-01-18T13:45:00Z',
          updated_at: '2024-01-20T12:15:00Z',
          last_login: '2024-01-20T12:15:00Z'
        }
      ];
      
      this.filterUsers();
    } catch (error) {
      this.errorHandler.handleError(error, 'Loading users');
    } finally {
      this.loading = false;
      this.loadingService.setLoading(loadingKey, false);
    }
  }

  filterUsers() {
    const term = this.searchTerm.toLowerCase();
    this.filteredUsers = this.users.filter(user =>
      user.username.toLowerCase().includes(term) ||
      user.email.toLowerCase().includes(term) ||
      user.realms.some(realm => realm.toLowerCase().includes(term))
    );
  }

  createUser() {
    this.router.navigate(['/users/create']);
  }

  viewUser(user: User) {
    this.router.navigate(['/users', user.id]);
  }

  editUser(user: User) {
    this.router.navigate(['/users', user.id, 'edit']);
  }

  managePermissions(user: User) {
    this.router.navigate(['/users', user.id, 'permissions']);
  }

  async toggleUserStatus(user: User, active: boolean) {
    this.updatingUser = user.id;
    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 500));
      
      user.active = !user.active;
      this.errorHandler.handleSuccess(
        `User ${user.username} ${user.active ? 'activated' : 'deactivated'} successfully`
      );
    } catch (error) {
      this.errorHandler.handleError(error, 'Updating user status');
    } finally {
      this.updatingUser = undefined;
    }
  }

  async deleteUser(user: User) {
    const confirmed = confirm(`Are you sure you want to delete user "${user.username}"?`);
    if (!confirmed) return;

    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 500));
      
      this.users = this.users.filter(u => u.id !== user.id);
      this.filterUsers();
      this.errorHandler.handleSuccess(`User ${user.username} deleted successfully`);
    } catch (error) {
      this.errorHandler.handleError(error, 'Deleting user');
    }
  }

  getRealmColor(realm: string): string {
    switch (realm.toLowerCase()) {
      case 'admins': return 'warn';
      case 'developers': return 'primary';
      case 'readonly': return 'accent';
      default: return '';
    }
  }

  getActiveUsersCount(): number {
    return this.users.filter(user => user.active).length;
  }

  getInactiveUsersCount(): number {
    return this.users.filter(user => !user.active).length;
  }

  getAdminUsersCount(): number {
    return this.users.filter(user => user.realms.includes('admins')).length;
  }
}
