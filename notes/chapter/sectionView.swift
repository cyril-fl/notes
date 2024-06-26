//
//  sectionView.swift
//  notes
//
//  Created by Cyril Flambard on 17/04/2024.
//

import SwiftUI

struct sectionView: View {
    
    var body: some View {
        HStack {
            
            ChapterView(title: "Titre1", content: "Content2")
                .thumbnail()
            ChapterView(title: "Titre2", content: "Content3")
                .thumbnail()
            ChapterView(title: "Titre3", content: "Content3")
                .thumbnail()
            
        }
    }
}

#Preview {
    sectionView()
}
