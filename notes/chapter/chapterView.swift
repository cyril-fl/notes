//
//  chapterView.swift
//  notes
//
//  Created by Cyril Flambard on 17/04/2024.
//

import SwiftUI


struct ChapterView: View {
    let id: UUID
    @State private var title: String
    @State private var content: String

    init(title: String, content: String) {
        self._title = State(initialValue: title)
        self._content = State(initialValue: content)
        self.id = UUID()
    }

    var body: some View {
        VStack(alignment: .leading) {
            titleText(title: title)
            bodyText(content: content)
            Spacer()
        }
    }
    
    
    func thumbnail() -> some View {
        VStack {

            Text(title)
                .font(.custom("Helvetica", size: 14, relativeTo: .body))
                .foregroundColor(.blue)
            Text(content)
                .font(.custom("Helvetica", size: 14, relativeTo: .body))
                .foregroundColor(.gray)
        }
        .frame(width: 100, height: 100)
        .background(.green)
        .cornerRadius(10)
    }
    

}





#Preview {
    ChapterView(title: "Mon titre", content: "Le contenue")
        .thumbnail()
}


