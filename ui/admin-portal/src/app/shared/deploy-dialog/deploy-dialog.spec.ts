import { ComponentFixture, TestBed } from '@angular/core/testing';

import { DeployDialog } from './deploy-dialog';

describe('DeployDialog', () => {
  let component: DeployDialog;
  let fixture: ComponentFixture<DeployDialog>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [DeployDialog]
    })
    .compileComponents();

    fixture = TestBed.createComponent(DeployDialog);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
