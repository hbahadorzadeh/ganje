import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, Router } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatTableModule } from '@angular/material/table';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatSnackBar } from '@angular/material/snack-bar';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatTabsModule } from '@angular/material/tabs';
import { MatChipsModule } from '@angular/material/chips';
import { GanjeApiService, ArtifactInfo, Repository } from '../../../core/services/ganje-api.service';
import { ErrorHandlerService } from '../../../core/services/error-handler.service';
import { LoadingService } from '../../../core/services/loading.service';
import { LoadingSpinnerComponent } from '../../../shared/components/loading-spinner/loading-spinner.component';
import { ErrorDisplayComponent } from '../../../shared/components/error-display/error-display.component';

@Component({
  selector: 'app-artifact-list',
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
    MatProgressBarModule,
    MatTooltipModule,
    MatTabsModule,
    MatChipsModule
  ],
  template: `
    <div class="artifact-container">
      <div class="page-header">
        <div class="header-content">
          <h1>{{ repository?.name }} Artifacts</h1>
          <p>Manage artifacts in this repository</p>
        </div>
        <button mat-raised-button color="primary" (click)="showUploadArea = !showUploadArea">
          <mat-icon>cloud_upload</mat-icon>
          {{ showUploadArea ? 'Hide Upload' : 'Upload Artifacts' }}
        </button>
      </div>

      <!-- Upload Area -->
      <mat-card *ngIf="showUploadArea" class="upload-card">
        <mat-card-header>
          <mat-card-title>
            <mat-icon>cloud_upload</mat-icon>
            Upload Artifacts
          </mat-card-title>
        </mat-card-header>
        <mat-card-content>
          <div class="upload-area" 
               [class.drag-over]="isDragOver"
               (dragover)="onDragOver($event)"
               (dragleave)="onDragLeave($event)"
               (drop)="onDrop($event)"
               (click)="fileInput.click()">
            <div class="upload-content">
              <mat-icon class="upload-icon">cloud_upload</mat-icon>
              <h3>Drag & Drop Files Here</h3>
              <p>or click to browse files</p>
              <p class="upload-hint">Supports all artifact types for {{ repository?.artifact_type | uppercase }}</p>
            </div>
            <input #fileInput type="file" multiple style="display: none" (change)="onFileSelected($event)">
          </div>

          <!-- Upload Queue -->
          <div *ngIf="uploadQueue.length > 0" class="upload-queue">
            <h4>Upload Queue</h4>
            <div *ngFor="let upload of uploadQueue; trackBy: trackUpload" class="upload-item">
              <div class="upload-info">
                <div class="file-details">
                  <strong>{{ upload.file.name }}</strong>
                  <span class="file-size">{{ formatBytes(upload.file.size) }}</span>
                </div>
                <div class="upload-path">
                  <mat-form-field appearance="outline" class="path-input">
                    <mat-label>Artifact Path</mat-label>
                    <input matInput [(ngModel)]="upload.path" placeholder="path/to/artifact">
                    <mat-hint>Specify the path where this artifact should be stored</mat-hint>
                  </mat-form-field>
                </div>
              </div>
              <div class="upload-progress">
                <mat-progress-bar 
                  [mode]="upload.status === 'uploading' ? 'determinate' : 'indeterminate'"
                  [value]="upload.progress"
                  [color]="getProgressColor(upload.status)">
                </mat-progress-bar>
                <div class="progress-text">
                  <span>{{ getStatusText(upload.status) }}</span>
                  <span *ngIf="upload.status === 'uploading'">{{ upload.progress }}%</span>
                </div>
              </div>
              <div class="upload-actions">
                <button mat-icon-button 
                        *ngIf="upload.status === 'pending'"
                        (click)="removeFromQueue(upload.id)"
                        matTooltip="Remove from queue">
                  <mat-icon>close</mat-icon>
                </button>
                <mat-icon *ngIf="upload.status === 'completed'" color="primary">check_circle</mat-icon>
                <mat-icon *ngIf="upload.status === 'error'" color="warn">error</mat-icon>
              </div>
            </div>
            
            <div class="queue-actions">
              <button mat-button (click)="clearQueue()" [disabled]="isUploading">
                Clear Queue
              </button>
              <button mat-raised-button color="primary" 
                      (click)="startUpload()" 
                      [disabled]="isUploading || !canStartUpload()">
                <mat-icon>cloud_upload</mat-icon>
                {{ isUploading ? 'Uploading...' : 'Start Upload' }}
              </button>
            </div>
          </div>
        </mat-card-content>
      </mat-card>

      <!-- Artifacts List -->
      <mat-card>
        <mat-card-header>
          <div class="card-header-content">
            <mat-card-title>Artifacts ({{ artifacts.length }})</mat-card-title>
            <div class="search-container">
              <mat-form-field appearance="outline">
                <mat-label>Search artifacts</mat-label>
                <input matInput [(ngModel)]="searchTerm" (input)="filterArtifacts()" placeholder="Search by name, version...">
                <mat-icon matSuffix>search</mat-icon>
              </mat-form-field>
            </div>
          </div>
        </mat-card-header>
        
        <mat-card-content>
          <div *ngIf="loading" class="loading-container">
            <mat-spinner></mat-spinner>
            <p>Loading artifacts...</p>
          </div>

          <table mat-table [dataSource]="filteredArtifacts" *ngIf="!loading && filteredArtifacts.length > 0">
            <ng-container matColumnDef="name">
              <th mat-header-cell *matHeaderCellDef>Name</th>
              <td mat-cell *matCellDef="let artifact">
                <div class="artifact-name">
                  <strong>{{ artifact.name }}</strong>
                  <div class="artifact-path">{{ artifact.path }}</div>
                </div>
              </td>
            </ng-container>

            <ng-container matColumnDef="version">
              <th mat-header-cell *matHeaderCellDef>Version</th>
              <td mat-cell *matCellDef="let artifact">
                <mat-chip-set>
                  <mat-chip>{{ artifact.version }}</mat-chip>
                </mat-chip-set>
              </td>
            </ng-container>

            <ng-container matColumnDef="size">
              <th mat-header-cell *matHeaderCellDef>Size</th>
              <td mat-cell *matCellDef="let artifact">{{ formatBytes(artifact.size) }}</td>
            </ng-container>

            <ng-container matColumnDef="checksum">
              <th mat-header-cell *matHeaderCellDef>Checksum</th>
              <td mat-cell *matCellDef="let artifact">
                <code class="checksum">{{ artifact.checksum.substring(0, 12) }}...</code>
              </td>
            </ng-container>

            <ng-container matColumnDef="created_at">
              <th mat-header-cell *matHeaderCellDef>Created</th>
              <td mat-cell *matCellDef="let artifact">{{ artifact.created_at | date:'short' }}</td>
            </ng-container>

            <ng-container matColumnDef="actions">
              <th mat-header-cell *matHeaderCellDef>Actions</th>
              <td mat-cell *matCellDef="let artifact">
                <div class="action-buttons">
                  <button mat-icon-button (click)="downloadArtifact(artifact)" matTooltip="Download">
                    <mat-icon>download</mat-icon>
                  </button>
                  <button mat-icon-button (click)="viewArtifactDetails(artifact)" matTooltip="View Details">
                    <mat-icon>info</mat-icon>
                  </button>
                  <button mat-icon-button color="warn" (click)="deleteArtifact(artifact)" matTooltip="Delete">
                    <mat-icon>delete</mat-icon>
                  </button>
                </div>
              </td>
            </ng-container>

            <tr mat-header-row *matHeaderRowDef="displayedColumns"></tr>
            <tr mat-row *matRowDef="let row; columns: displayedColumns;"></tr>
          </table>

          <div *ngIf="!loading && filteredArtifacts.length === 0" class="empty-state">
            <mat-icon>inventory_2</mat-icon>
            <h3>No artifacts found</h3>
            <p *ngIf="searchTerm">Try adjusting your search criteria</p>
            <p *ngIf="!searchTerm">Upload your first artifact to get started</p>
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
    .artifact-container {
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

    .upload-card {
      margin-bottom: 2rem;
    }

    .upload-area {
      border: 2px dashed #ccc;
      border-radius: 8px;
      padding: 3rem;
      text-align: center;
      cursor: pointer;
      transition: all 0.3s ease;
      background: #fafafa;
    }

    .upload-area:hover,
    .upload-area.drag-over {
      border-color: #3f51b5;
      background: #f3f4ff;
    }

    .upload-content {
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 1rem;
    }

    .upload-icon {
      font-size: 4rem;
      width: 4rem;
      height: 4rem;
      color: #666;
    }

    .upload-hint {
      font-size: 0.875rem;
      color: #666;
    }

    .upload-queue {
      margin-top: 2rem;
      padding-top: 2rem;
      border-top: 1px solid #e0e0e0;
    }

    .upload-item {
      display: flex;
      align-items: center;
      gap: 1rem;
      padding: 1rem;
      border: 1px solid #e0e0e0;
      border-radius: 4px;
      margin-bottom: 1rem;
    }

    .upload-info {
      flex: 1;
    }

    .file-details {
      display: flex;
      align-items: center;
      gap: 1rem;
      margin-bottom: 0.5rem;
    }

    .file-size {
      font-size: 0.875rem;
      color: #666;
    }

    .path-input {
      width: 100%;
    }

    .upload-progress {
      flex: 0 0 200px;
    }

    .progress-text {
      display: flex;
      justify-content: space-between;
      font-size: 0.875rem;
      margin-top: 0.25rem;
    }

    .upload-actions {
      flex: 0 0 auto;
    }

    .queue-actions {
      display: flex;
      gap: 1rem;
      justify-content: flex-end;
      margin-top: 1rem;
      padding-top: 1rem;
      border-top: 1px solid #e0e0e0;
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

    .artifact-name {
      display: flex;
      flex-direction: column;
    }

    .artifact-path {
      font-size: 0.875rem;
      color: #666;
      font-family: monospace;
    }

    .checksum {
      font-family: monospace;
      font-size: 0.875rem;
      background: #f5f5f5;
      padding: 0.25rem 0.5rem;
      border-radius: 4px;
    }

    .action-buttons {
      display: flex;
      gap: 0.5rem;
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
      
      .upload-item {
        flex-direction: column;
        align-items: stretch;
      }
      
      .upload-progress {
        flex: auto;
      }
    }
  `]
})
export class ArtifactListComponent implements OnInit {
  repository?: Repository;
  artifacts: ArtifactInfo[] = [];
  filteredArtifacts: ArtifactInfo[] = [];
  loading = true;
  errorMessage = '';
  searchTerm = '';
  showUploadArea = false;
  
  // Upload functionality
  uploadQueue: UploadItem[] = [];
  isUploading = false;
  isDragOver = false;
  
  displayedColumns: string[] = ['name', 'version', 'size', 'checksum', 'created_at', 'actions'];

  constructor(
    private ganjeApiService: GanjeApiService,
    private route: ActivatedRoute,
    private router: Router,
    private snackBar: MatSnackBar,
    private errorHandler: ErrorHandlerService,
    private loadingService: LoadingService
  ) {}

  async ngOnInit() {
    const repositoryName = this.route.snapshot.paramMap.get('name');
    if (repositoryName) {
      await this.loadRepository(repositoryName);
      await this.loadArtifacts(repositoryName);
    }
  }

  private async loadRepository(name: string) {
    try {
      this.repository = await this.ganjeApiService.getRepository(name).toPromise();
    } catch (error: any) {
      this.errorMessage = 'Failed to load repository information';
      this.errorHandler.handleError(error, 'Loading repository');
    }
  }

  private async loadArtifacts(repositoryName: string) {
    const loadingKey = 'artifacts-list';
    try {
      this.loading = true;
      this.errorMessage = '';
      this.loadingService.setLoading(loadingKey, true);
      this.artifacts = await this.ganjeApiService.getArtifacts(repositoryName).toPromise() || [];
      this.filterArtifacts();
    } catch (error: any) {
      this.errorMessage = 'Failed to load artifacts';
      this.errorHandler.handleError(error, 'Loading artifacts');
    } finally {
      this.loading = false;
      this.loadingService.setLoading(loadingKey, false);
    }
  }

  filterArtifacts() {
    const term = this.searchTerm.toLowerCase();
    this.filteredArtifacts = this.artifacts.filter(artifact =>
      artifact.name.toLowerCase().includes(term) ||
      artifact.version.toLowerCase().includes(term) ||
      artifact.path.toLowerCase().includes(term)
    );
  }

  // Drag and drop handlers
  onDragOver(event: DragEvent) {
    event.preventDefault();
    this.isDragOver = true;
  }

  onDragLeave(event: DragEvent) {
    event.preventDefault();
    this.isDragOver = false;
  }

  onDrop(event: DragEvent) {
    event.preventDefault();
    this.isDragOver = false;
    
    const files = event.dataTransfer?.files;
    if (files) {
      this.addFilesToQueue(Array.from(files));
    }
  }

  onFileSelected(event: any) {
    const files = event.target.files;
    if (files) {
      this.addFilesToQueue(Array.from(files));
    }
    // Reset the input
    event.target.value = '';
  }

  private addFilesToQueue(files: File[]) {
    files.forEach(file => {
      const uploadItem: UploadItem = {
        id: Date.now() + Math.random(),
        file,
        path: this.generateDefaultPath(file),
        status: 'pending',
        progress: 0
      };
      this.uploadQueue.push(uploadItem);
    });
  }

  private generateDefaultPath(file: File): string {
    // Generate a default path based on artifact type and file name
    const artifactType = this.repository?.artifact_type || 'generic';
    
    switch (artifactType) {
      case 'maven':
        return `com/example/${file.name}`;
      case 'npm':
        return `@scope/${file.name}`;
      case 'docker':
        return `library/${file.name}`;
      default:
        return file.name;
    }
  }

  trackUpload(index: number, item: UploadItem): any {
    return item.id;
  }

  removeFromQueue(id: number) {
    this.uploadQueue = this.uploadQueue.filter(item => item.id !== id);
  }

  clearQueue() {
    this.uploadQueue = this.uploadQueue.filter(item => item.status === 'uploading');
  }

  canStartUpload(): boolean {
    return this.uploadQueue.some(item => item.status === 'pending' && item.path.trim());
  }

  async startUpload() {
    if (!this.repository || this.isUploading) return;

    this.isUploading = true;
    const pendingUploads = this.uploadQueue.filter(item => item.status === 'pending');

    for (const upload of pendingUploads) {
      if (!upload.path.trim()) {
        upload.status = 'error';
        continue;
      }

      try {
        upload.status = 'uploading';
        upload.progress = 0;

        // Simulate progress updates
        const progressInterval = setInterval(() => {
          if (upload.progress < 90) {
            upload.progress += Math.random() * 20;
          }
        }, 200);

        await this.ganjeApiService.uploadArtifact(this.repository.name, upload.path, upload.file).toPromise();
        
        clearInterval(progressInterval);
        upload.progress = 100;
        upload.status = 'completed';
        
        this.errorHandler.handleSuccess(`${upload.file.name} uploaded successfully`);
      } catch (error: any) {
        upload.status = 'error';
        this.errorHandler.handleError(error, `Uploading ${upload.file.name}`);
      }
    }

    this.isUploading = false;
    
    // Refresh artifacts list
    if (this.repository) {
      await this.loadArtifacts(this.repository.name);
    }
  }

  getProgressColor(status: string): string {
    switch (status) {
      case 'completed': return 'primary';
      case 'error': return 'warn';
      default: return 'primary';
    }
  }

  getStatusText(status: string): string {
    switch (status) {
      case 'pending': return 'Ready to upload';
      case 'uploading': return 'Uploading';
      case 'completed': return 'Completed';
      case 'error': return 'Failed';
      default: return 'Unknown';
    }
  }

  async downloadArtifact(artifact: ArtifactInfo) {
    if (!this.repository) return;

    try {
      const blob = await this.ganjeApiService.downloadArtifact(this.repository.name, artifact.path).toPromise();
      if (blob) {
        const url = window.URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        link.download = artifact.name;
        link.click();
        window.URL.revokeObjectURL(url);
      }
    } catch (error: any) {
      this.errorHandler.handleError(error, 'Downloading artifact');
    }
  }

  viewArtifactDetails(artifact: ArtifactInfo) {
    // Navigate to artifact details page
    this.router.navigate(['/repositories', this.repository?.name, 'artifacts', artifact.id]);
  }

  async deleteArtifact(artifact: ArtifactInfo) {
    if (!this.repository) return;

    const confirmed = confirm(`Are you sure you want to delete "${artifact.name}"?`);
    if (!confirmed) return;

    try {
      await this.ganjeApiService.deleteArtifact(this.repository.name, artifact.path).toPromise();
      this.errorHandler.handleSuccess('Artifact deleted successfully');
      await this.loadArtifacts(this.repository.name);
    } catch (error: any) {
      this.errorHandler.handleError(error, 'Deleting artifact');
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

interface UploadItem {
  id: number;
  file: File;
  path: string;
  status: 'pending' | 'uploading' | 'completed' | 'error';
  progress: number;
}
