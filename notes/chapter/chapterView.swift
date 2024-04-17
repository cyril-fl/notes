//
//  chapterView.swift
//  notes
//
//  Created by Cyril Flambard on 17/04/2024.
//

import SwiftUI

struct chapterView: View {
    @State private var title: String
    @State private var content: String
    
    
    init(title: String, content: String) {
        self._title = State(initialValue: title)
        self._content = State(initialValue: content)
    }
    
    
    var body: some View {
        titleText(title: title)
        bodyText(content: content)
    }
}

#Preview {
    chapterView(title: "Mon titre", content: "Le contenue")
}
