import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormBuilder, Validators, FormGroup } from '@angular/forms';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatDialogModule, MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { Inject } from '@angular/core';
import { ArtifactsService } from '../../core/services/artifacts';

@Component({
  selector: 'app-deploy-dialog',
  imports: [CommonModule, ReactiveFormsModule, MatFormFieldModule, MatInputModule, MatButtonModule, MatDialogModule],
  templateUrl: './deploy-dialog.html',
  styleUrl: './deploy-dialog.scss'
})
export class DeployDialog {
  form!: FormGroup;

  constructor(
    private fb: FormBuilder,
    private artifacts: ArtifactsService,
    private dialogRef: MatDialogRef<DeployDialog>,
    @Inject(MAT_DIALOG_DATA) public data: { repo: string }
  ) {
    this.form = this.fb.group({
      path: ['', Validators.required],
      file: [null as File | null, Validators.required],
    });
  }

  onFileChange(evt: Event) {
    const input = evt.target as HTMLInputElement;
    const file = input.files && input.files.length ? input.files[0] : null;
    this.form.patchValue({ file });
  }

  submit() {
    if (this.form.invalid || !this.form.value.file) return;
    const { path, file } = this.form.value;
    this.artifacts.uploadMultipart(this.data.repo, String(path), file as File).subscribe({
      next: (res) => this.dialogRef.close(res),
      error: (err) => this.dialogRef.close({ error: err })
    });
  }

  cancel() {
    this.dialogRef.close();
  }
}
