import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatTableModule } from '@angular/material/table';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSnackBar } from '@angular/material/snack-bar';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatDialog } from '@angular/material/dialog';
import { GanjeApiService, Repository } from '../../../core/services/ganje-api.service';
import { ErrorHandlerService } from '../../../core/services/error-handler.service';
import { LoadingService } from '../../../core/services/loading.service';
import { LoadingSpinnerComponent } from '../../../shared/components/loading-spinner/loading-spinner.component';
import { ErrorDisplayComponent } from '../../../shared/components/error-display/error-display.component';
import { RepositoryDeleteDialogComponent } from './repository-delete-dialog.component';

@Component({
  selector: 'app-repository-list',
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
    MatTooltipModule,
    LoadingSpinnerComponent,
    ErrorDisplayComponent
  ],
  template: `
    <div class="repository-list-container">
      <div class="page-header">
        <div class="header-content">
          <h1>Repositories</h1>
          <p>Manage your artifact repositories</p>
        </div>
        <div class="header-actions">
          <button mat-raised-button color="primary" (click)="createRepository()">
            <mat-icon>add</mat-icon>
            Create Repository
          </button>
        </div>
      </div>

      <mat-card>
        <mat-card-header>
          <mat-card-title>Repository List</mat-card-title>
        </mat-card-header>

        <mat-card-content>
          <div class="search-container">
            <mat-form-field appearance="outline" class="search-field">
              <mat-label>Search repositories</mat-label>
              <input matInput 
                     placeholder="Search repositories..." 
                     [(ngModel)]="searchTerm"
                     (input)="filterRepositories()">
              <mat-icon matSuffix>search</mat-icon>
            </mat-form-field>
          </div>

          <app-loading-spinner 
            *ngIf="loading" 
            message="Loading repositories...">
          </app-loading-spinner>

          <app-error-display 
            *ngIf="errorMessage" 
            [message]="errorMessage"
            title="Failed to Load Repositories"
            [retryAction]="loadRepositories.bind(this)">
          </app-error-display>

          <div *ngIf="!loading && filteredRepositories.length === 0 && !errorMessage" class="empty-state">
            <mat-icon class="empty-icon">archive</mat-icon>
            <h4>No repositories found</h4>
            <p>Create your first repository to get started</p>
            <button mat-raised-button color="primary" (click)="createRepository()">
              <mat-icon>add</mat-icon>
              Create Repository
            </button>
          </div>

          <table mat-table [dataSource]="filteredRepositories" *ngIf="!loading && filteredRepositories.length > 0" class="repository-table">
            <ng-container matColumnDef="name">
              <th mat-header-cell *matHeaderCellDef>Name</th>
              <td mat-cell *matCellDef="let repo">
                <div class="repo-name-cell">
                  <strong>{{ repo.name }}</strong>
                  <small *ngIf="repo.url" class="repo-url">{{ repo.url }}</small>
                </div>
              </td>
            </ng-container>

            <ng-container matColumnDef="type">
              <th mat-header-cell *matHeaderCellDef>Type</th>
              <td mat-cell *matCellDef="let repo">
                <span class="type-badge">{{ repo.type }}</span>
              </td>
            </ng-container>

            <ng-container matColumnDef="artifactType">
              <th mat-header-cell *matHeaderCellDef>Artifact Type</th>
              <td mat-cell *matCellDef="let repo">
                <span class="artifact-type-badge">{{ repo.artifact_type }}</span>
              </td>
            </ng-container>

            <ng-container matColumnDef="artifacts">
              <th mat-header-cell *matHeaderCellDef>Artifacts</th>
              <td mat-cell *matCellDef="let repo">{{ repo.total_artifacts | number }}</td>
            </ng-container>

            <ng-container matColumnDef="size">
              <th mat-header-cell *matHeaderCellDef>Size</th>
              <td mat-cell *matCellDef="let repo">{{ formatBytes(repo.total_size) }}</td>
            </ng-container>

            <ng-container matColumnDef="actions">
              <th mat-header-cell *matHeaderCellDef>Actions</th>
              <td mat-cell *matCellDef="let repo">
                <div class="action-buttons">
                  <button mat-icon-button (click)="viewRepository(repo.name)" matTooltip="View Details">
                    <mat-icon>visibility</mat-icon>
                  </button>
                  <button mat-icon-button (click)="editRepository(repo.name)" matTooltip="Edit Repository">
                    <mat-icon>edit</mat-icon>
                  </button>
                  <button mat-icon-button color="warn" (click)="deleteRepository(repo)" matTooltip="Delete Repository">
                    <mat-icon>delete</mat-icon>
                  </button>
                </div>
              </td>
            </ng-container>

            <tr mat-header-row *matHeaderRowDef="displayedColumns"></tr>
            <tr mat-row *matRowDef="let row; columns: displayedColumns;"></tr>
          </table>

          <div *ngIf="errorMessage" class="error-message">
            <mat-icon color="warn">error</mat-icon>
            {{ errorMessage }}
          </div>
        </mat-card-content>
      </mat-card>
    </div>
  `,
  styles: [`
    .repository-list-container {
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
      color: var(--text-basic-color);
    }

    .header-content p {
      margin: 0;
      color: var(--text-hint-color);
    }

    .card-header-content {
      display: flex;
      justify-content: space-between;
      align-items: center;
      width: 100%;
    }

    .card-header-content h6 {
      margin: 0;
    }

    .search-container {
      width: 300px;
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

    .empty-state {
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      padding: 4rem 2rem;
      text-align: center;
    }

    .empty-icon {
      font-size: 4rem;
      color: var(--text-hint-color);
      margin-bottom: 1rem;
    }

    .empty-state h4 {
      margin: 0 0 0.5rem;
      color: var(--text-basic-color);
    }

    .empty-state p {
      margin: 0 0 2rem;
      color: var(--text-hint-color);
    }

    .repository-table {
      width: 100%;
    }

    .repository-row:hover {
      background-color: var(--background-basic-color-2);
    }

    .repo-name-cell {
      display: flex;
      flex-direction: column;
      gap: 0.25rem;
    }

    .repo-url {
      color: var(--text-hint-color);
      font-size: 0.75rem;
    }

    .action-buttons {
      display: flex;
      gap: 0.5rem;
    }

    @media (max-width: 768px) {
      .repository-list-container {
        padding: 1rem;
      }

      .page-header {
        flex-direction: column;
        gap: 1rem;
        align-items: stretch;
      }

      .card-header-content {
        flex-direction: column;
        gap: 1rem;
        align-items: stretch;
      }

      .search-container {
        width: 100%;
      }

      .repository-table {
        font-size: 0.875rem;
      }

      .action-buttons {
        flex-direction: column;
      }
    }
  `]
})
export class RepositoryListComponent implements OnInit {
  repositories: Repository[] = [];
  filteredRepositories: Repository[] = [];
  loading = true;
  errorMessage = '';
  searchTerm = '';
  displayedColumns: string[] = ['name', 'type', 'artifact_type', 'total_artifacts', 'total_size', 'created_at', 'actions'];

  constructor(
    private ganjeApiService: GanjeApiService,
    private router: Router,
    private snackBar: MatSnackBar,
    private dialog: MatDialog,
    private errorHandler: ErrorHandlerService,
    private loadingService: LoadingService
  ) {}

  async ngOnInit() {
    await this.loadRepositories();
  }

  async loadRepositories() {
    const loadingKey = 'repositories-list';
    try {
      this.loading = true;
      this.errorMessage = '';
      this.loadingService.setLoading(loadingKey, true);
      
      this.repositories = await this.ganjeApiService.getRepositories().toPromise() || [];
      this.filteredRepositories = [...this.repositories];
    } catch (error) {
      this.errorMessage = 'Failed to load repositories. Please try again.';
      this.errorHandler.handleError(error, 'Loading repositories');
    } finally {
      this.loading = false;
      this.loadingService.setLoading(loadingKey, false);
    }
  }

  filterRepositories() {
    if (!this.searchTerm.trim()) {
      this.filteredRepositories = [...this.repositories];
      return;
    }

    const term = this.searchTerm.toLowerCase();
    this.filteredRepositories = this.repositories.filter(repo =>
      repo.name.toLowerCase().includes(term) ||
      repo.type.toLowerCase().includes(term) ||
      repo.artifact_type.toLowerCase().includes(term)
    );
  }

  createRepository() {
    this.router.navigate(['/repositories/create']);
  }

  editRepository(repositoryName: string) {
    this.router.navigate(['/repositories/edit', repositoryName]);
  }

  viewRepository(repositoryName: string) {
    this.router.navigate(['/repositories/detail', repositoryName]);
  }

  deleteRepository(repository: Repository) {
    const dialogRef = this.dialog.open(RepositoryDeleteDialogComponent, {
      width: '600px',
      data: { repository }
    });

    dialogRef.afterClosed().subscribe(async (result) => {
      if (result?.confirmed) {
        try {
          await this.ganjeApiService.deleteRepository(repository.name, result.force).toPromise();
          this.errorHandler.handleSuccess(`Repository "${repository.name}" deleted successfully`);
          await this.loadRepositories(); // Refresh the list
        } catch (error: any) {
          this.errorHandler.handleError(error, 'Deleting repository');
        }
      }
    });
  }


  getTypeStatus(type: string): string {
    switch (type.toLowerCase()) {
      case 'local':
        return 'success';
      case 'remote':
        return 'warning';
      case 'virtual':
        return 'info';
      default:
        return 'basic';
    }
  }

  formatBytes(bytes: number): string {
    if (bytes === 0) return '0 Bytes';
    
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }
}
