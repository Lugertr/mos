import { ChangeDetectionStrategy, Component, effect, input } from '@angular/core';
import { MatTableModule } from '@angular/material/table';
import { TableDataSource } from '@core/table-data-source/table-data-source.class';
import { ScrollContainerComponent } from '@core/scroll-container/scroll-container.component';
import { DatePipe } from '@angular/common';
import { ArchiveDocument } from '../services/dashboard.service';
@Component({
  selector: 'app-dashboard-table',
  templateUrl: './dashboard-table.component.html',
  styleUrl: './dashboard-table.component.scss',
  standalone: true,
  imports: [DatePipe, MatTableModule, ScrollContainerComponent],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class DashboardTableComponent {
  documents = input<ArchiveDocument[]>([]);
  readonly dataSource = new TableDataSource<ArchiveDocument>([]);
  readonly displayedColumns = ['doc_id', 'title', 'type_name', 'author_name', 'document_date', 'privacy'];

  constructor() {
    effect(() => {
      this.dataSource.data = this.documents();
      console.log('trigger');
    })
  }
}
