//
//  thumbnailsAdd.swift
//  notes
//
//  Created by Cyril Flambard on 18/04/2024.
//

import SwiftUI

struct thumbnailsAdd: View {
    
    @State private var isPressed : Bool
    
    init() {
        self._isPressed = State(initialValue: false)
    }


    var body: some View {
        HStack {
            Text("+")
                .frame(width: 100, height: 100)
                .background(isPressed ? Color.red : Color.blue)
                .cornerRadius(10)
                .onTapGesture {
                    isPressed.toggle()
                    
                }
//                .onHover(perform: { hovering in
//                    if hovering {
//                        self.background(TintShapeStyle(.red)
//                    } else {
//                        self.background(.clear)
//                    }
//            })                .background(isPressed ? Color.red : Color.blue)
        }
    }
}

#Preview {
    thumbnailsAdd()
}
