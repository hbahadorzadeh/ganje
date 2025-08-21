import { ComponentFixture, TestBed } from '@angular/core/testing';

import { MoveDialog } from './move-dialog';

describe('MoveDialog', () => {
  let component: MoveDialog;
  let fixture: ComponentFixture<MoveDialog>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [MoveDialog]
    })
    .compileComponents();

    fixture = TestBed.createComponent(MoveDialog);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
